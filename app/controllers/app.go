package controllers

import (
    "github.com/robfig/revel"
	"net/http"
	"net/url"
	"io"
	"io/ioutil"
	"os/exec"
	"os"
	"strings"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) About() revel.Result {
	return c.Render()
}

func (c App) Result(pdf string) revel.Result {
    
	c.Validation.Required(pdf).Message("You must provide a url to a pdf.")
	
    if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}
	
	url, err := url.Parse(pdf)
	if err != nil {
	    c.Validation.Error("Error with PDF URL", err)
    }
	
	//TODO: messy. Maybe defer? Can defer contains returns?
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}
	
	if url.Scheme == ""{
	    url.Scheme = "http"
	}
	
	resp, err := http.Get(url.String())
    if err != nil {
	    c.Validation.Error("HTTP Error with PDF URL", err)
    }
	
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}	
	
	if resp.StatusCode != http.StatusOK  {
	    c.Validation.Error("Error loading pdf file.", resp.Status)
	}
	
	if resp.Header["Content-Type"][0] != "application/pdf" {
	    c.Validation.Error("Error loading pdf file, server not returning pdf content.")
	}
	
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}
	
	defer resp.Body.Close()
	
	//Create a temporary file, copy the body into it. Then we can run preflight.
	temp, err := ioutil.TempFile("", "pdfserver")
	if err != nil{
	    c.Validation.Error("Error creating temp file.", err)
	}
	defer os.Remove(temp.Name())
	
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}
	
	_, err = io.Copy(temp, resp.Body)	
	
	if err != nil{
	    c.Validation.Error("Error writing to temp file.", err)
	}
	
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}
	
	out, _ := exec.Command("java", "-jar", "preflight-app-1.8.2.jar", temp.Name()).Output()

	outSlice := strings.Split(string(out), "\n")
	return c.Render(pdf, outSlice)
}