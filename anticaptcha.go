package anticaptcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

var (
	baseURL = &url.URL{Host: "api.anti-captcha.com", Scheme: "https", Path: "/"}
)

type LogFunction func(params ...interface{})
type Client struct {
	APIKey       string
	SendInterval time.Duration
	LogFunction  LogFunction
}

// Method to create the task to process the recaptcha, returns the task_id
func (c *Client) createTaskRecaptcha(websiteURL string, recaptchaKey string) (float64, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type":       "NoCaptchaTaskProxyless",
			"websiteURL": websiteURL,
			"websiteKey": recaptchaKey,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)

	if errorCode, ok := responseBody["errorCode"].(string); ok {
		return 0, errors.New(errorCode)
	}
	if taskId, ok := responseBody["taskId"].(float64); ok {
		return taskId, nil
	} else {
		return 0, errors.New("task id is empty")
	}
}

// Method to check the result of a given task, returns the json returned from the api
func (c *Client) getTaskResult(taskID float64) (map[string]interface{}, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"taskId":    taskID,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/getTaskResult"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	return responseBody, nil
}

// SendRecaptcha Method to encapsulate the processing of the recaptcha
// Given a url and a key, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendRecaptcha(websiteURL string, recaptchaKey string) (string, error) {
	// Create the task on anti-captcha api and get the task_id
	taskID, err := c.createTaskRecaptcha(websiteURL, recaptchaKey)
	if err != nil {
		return "", err
	}

	// Check if the result is ready, if not loop until it is
	response, err := c.getTaskResult(taskID)
	if err != nil {
		return "", err
	}
	for {
		if response["status"] == "processing" {
			c.log("Result is not ready, waiting a few seconds to check again...")
			time.Sleep(c.SendInterval)
			response, err = c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
		} else {
			c.log("Result is ready.")
			break
		}
	}
	return response["solution"].(map[string]interface{})["gRecaptchaResponse"].(string), nil
}

func (c Client) log(params ...interface{}) {
	if c.LogFunction != nil {
		c.LogFunction(params...)
	}
}

// Method to create the task to process the image captcha, returns the task_id
func (c *Client) createTaskImage(imgString string) (float64, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type": "ImageToTextTask",
			"body": imgString,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)

	if errorCode, ok := responseBody["errorCode"].(string); ok {
		return 0, errors.New(errorCode)
	}
	if taskId, ok := responseBody["taskId"].(float64); ok {
		return taskId, nil
	} else {
		return 0, errors.New("task id is empty")
	}
}

// SendImage Method to encapsulate the processing of the image captcha
// Given a base64 string from the image, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendImage(imgString string) (string, error) {
	// Create the task on anti-captcha api and get the task_id
	taskID, err := c.createTaskImage(imgString)
	if err != nil {
		return "", err
	}

	// Check if the result is ready, if not loop until it is
	response, err := c.getTaskResult(taskID)
	if err != nil {
		return "", err
	}
	for {
		if response["status"] == "processing" {
			c.log("Result is not ready, waiting a few seconds to check again...")
			time.Sleep(c.SendInterval)
			response, err = c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
		} else {
			c.log("Result is ready.")
			break
		}
	}
	return response["solution"].(map[string]interface{})["text"].(string), nil
}
