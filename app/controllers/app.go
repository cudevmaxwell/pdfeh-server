package controllers

import (
    "github.com/robfig/revel"
	"io/ioutil"
	"strings"
	"encoding/json"
	"net/url"
)

func init() {
	revel.TemplateFuncs["replace"] = strings.Replace
}

type App struct {
	*revel.Controller
}

type PDFSchemaValidatorList struct {
    Categories []PDFASchemaValidationErrorCategory
}

type PDFASchemaValidationErrorCategory struct {
    Errors []PDFASchemaValidationError
	Name string
	SubCategories []PDFASchemaValidationErrorCategory
}

type PDFASchemaValidationError struct {
	Code string
	Label string
}

func (c App) Index() revel.Result {

	return c.Render()
}

func (c App) Editor() revel.Result {

    file, err := ioutil.ReadFile("./errors.json")
    if err != nil {
	    c.Validation.Error("Cannot open validation errors json file", err)
    }
	
	//TODO: messy. Maybe defer? Can defer contains returns?
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Render()
	}
	
	
    var errorList PDFSchemaValidatorList
    err = json.Unmarshal(file, &errorList)
	
	if err != nil {
	    c.Validation.Error("Error in json file.", err)
    }
	
	//TODO: messy. Maybe defer? Can defer contains returns?
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Render()
	}

	return c.Render(errorList)
}

func (c App) About() revel.Result {
	return c.Render()
}

func (c App) Example() revel.Result {
	return c.Render()
}

func (c App) Result(pdf, validator string) revel.Result {
  
	c.Validation.Required(pdf).Message("You must provide a url to a pdf.")

    if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}
	
	v := url.Values{}
	v.Set("pdf",pdf)
	v.Set("validator", validator)	
	
	return c.Redirect("/api?%s", v.Encode())
}