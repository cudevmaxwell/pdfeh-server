package controllers

import (
    "github.com/robfig/revel"
    "strings"	
	"net/url"
	"net/http"
	"io"
	"io/ioutil"
	"os/exec"
	"os"
	"encoding/json"
)

type Api struct {
	*revel.Controller
}

func addHeaderCORS(c *revel.Controller) revel.Result {
	c.Response.Out.Header().Add("Access-Control-Allow-Origin","*")
	return nil
}

func init() {
    revel.InterceptFunc(addHeaderCORS, revel.AFTER, &Api{})
	revel.TemplateFuncs["replace"] = strings.Replace
}

type Validator struct {
    Entries []ValidatorEntry
}

type ValidatorEntry struct {
    Code string
	Level string
}

type Result struct {
    URL string
    Level string
	ValidationErrors []ValidationError
}

type ValidationError struct {
    Code string
	Text string
	Level string
    NumberOfTimes int	
}

type ResultError struct {
    Error string
}

func (c Api) PublicPDFPublicValidator(pdf, validator string) revel.Result {

    error := *c.Validation.Required(pdf)
	if !error.Ok {
	    return c.RenderJson(ResultError{"You must provide a url to a pdf."})
	}
	
	pdfUrl, err := url.Parse(pdf)
	if err != nil {
	    return c.RenderJson(ResultError{"Error with PDF URL."})
    }
		
	if pdfUrl.Scheme == "" {
	    pdfUrl.Scheme = "http"
	}
	
	resp, err := http.Get(pdfUrl.String())
    if err != nil {
	    return c.RenderJson(ResultError{"Error GETing PDF."})
    }
	
	if resp.StatusCode != http.StatusOK  {
	    return c.RenderJson(ResultError{"Error loading PDF."})
	}
	
	if resp.Header["Content-Type"][0] != "application/pdf" {
	    return c.RenderJson(ResultError{"Error loading pdf file, server not returning pdf content."})
	}
	
	defer resp.Body.Close()
	
	//Create a temporary file, copy the body into it. Then we can run preflight.
	temp, err := ioutil.TempFile("", "pdfserver")
	if err != nil{
	    return c.RenderJson(ResultError{"Error creating temp file."})
	}
	defer os.Remove(temp.Name())
	
	_, err = io.Copy(temp, resp.Body)	
	
	if err != nil{
	    return c.RenderJson(ResultError{"Error writing to temp file."})
	}
	
	validatorToUse := new(Validator)
	
	error = *c.Validation.Required(validator)
	
	if error.Ok {	

	    validationUrl, err := url.Parse(validator)
	    if err != nil {
	        return c.RenderJson(ResultError{"Error with validator url."})
        }
		    
	    if validationUrl.Scheme == "" {
	        validationUrl.Scheme = "http"
	    }		
	    
	    validatorResult, err := http.Get(validationUrl.String())
        if err != nil {
	        return c.RenderJson(ResultError{"HTTP Error with validator url."})
        }
	    
	    
	    if validatorResult.StatusCode != http.StatusOK  {
	        return c.RenderJson(ResultError{"Error loading validator url."})
	    }
	    	
	    defer validatorResult.Body.Close()
		
		jsonReadIn, err := ioutil.ReadAll(validatorResult.Body)
		
		if err != nil {
	        return c.RenderJson(ResultError{"Couldn't read the JSON document."})
        }

        err = json.Unmarshal(jsonReadIn, validatorToUse)	
        if err != nil {
	        return c.RenderJson(ResultError{"Couldn't unmarshal the JSON document." + err.Error()})
        }		
	}
		
	result := new(Result)
	
	result.Level = "pass"
	
	result.URL = pdf
	
	out, _ := exec.Command("java", "-jar", "preflight-app-1.8.2.jar", temp.Name()).Output()
	
	outputMap := make(map[string]int)	
	
	outSlice := strings.Split(string(out), "\n")[1:]
	
	for _, line := range outSlice {
	    elem, present := outputMap[line]
		if present == true {
			outputMap[line] = elem+1
		} else if strings.TrimSpace(line) != "" {		
            outputMap[line] = 1
		}
    }
	
	for error, numberOfTimes := range outputMap {
	    code := strings.TrimSpace(strings.Split(error, ":")[0])
		text := strings.TrimSpace(strings.Split(error, ":")[1])
		level := "pass"
		for _, validatorEntry := range validatorToUse.Entries{
		   if code == validatorEntry.Code  {
		     level = validatorEntry.Level
		   }
		}
	    result.ValidationErrors = append(result.ValidationErrors, ValidationError{code,text,level,numberOfTimes})
	}
	
	for _, validatorError := range result.ValidationErrors {
		if validatorError.Level == "fail"  {
		   result.Level = "fail"
		}  else if validatorError.Level == "warn" && result.Level == "pass"  {
		   result.Level = "warn"
		}
	}
	
	
	
	return c.RenderJson(result)
}