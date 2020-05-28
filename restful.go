package comms

import (
  "net/http"
  "io/ioutil"
  "encoding/json"
  "fmt"
  "bytes"
  "errors"
)

// res: pointer to the instance of response struct
func SendGetRequest(url string, res interface{}) error {
  response, err := http.Get(url)
  if err != nil {
    return err
	}
  if response.StatusCode != 200 {return errors.New(fmt.Sprintf("Error response code:%d",response.StatusCode))}
  body, err := ioutil.ReadAll(response.Body)
  if err != nil {
    return err
  }
  err = json.Unmarshal(body, res)
  if err != nil {
    return err
  }
  response.Body.Close()
  return nil
}

// sendBody: request body
// res: pointer to the instance of response struct
func SendPostRequest(url string, sendBody interface{}, res interface{}) error {
  requestBody, err := json.Marshal(sendBody)
  if err != nil {
    return err
	}
  response, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
  if err != nil {
    return err
	}
  if response.StatusCode != 200 {return errors.New(fmt.Sprintf("Error response code:%d",response.StatusCode))}
  // need response info from this post request
  if res != nil {
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
      return err
    }
    err = json.Unmarshal(body, res)
    if err != nil {
      return err
    }
    response.Body.Close()
  }
  return nil
}
