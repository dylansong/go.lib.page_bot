package page_bot

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestSendGetAllDeployments(t *testing.T) {
	// Send request and get response
	resp, err := sendGetAllDeployments()
	if err != nil {
		t.Errorf("Failed to send request: %v", err)
	}

	// Read response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response body: %v", err)
	}

	fmt.Printf("response Body : %s\n", string(respBody))

	//// Decode JSON response
	//var deployments []struct {
	//	ID          string `json:"id"`
	//	Name        string `json:"name"`
	//	URL         string `json:"url"`
	//	Status      string `json:"status"`
	//	CreatedOn   string `json:"created_on"`
	//	ModifiedOn  string `json:"modified_on"`
	//	Description string `json:"description"`
	//}
	//
	//if err := json.Unmarshal(respBody, &deployments); err != nil {
	//	t.Errorf("Failed to decode JSON response: %v", err)
	//}
	//
	//// Print deployment IDs and names
	//for _, deployment := range deployments {
	//	fmt.Printf("ID: %s, Name: %s\n", deployment.ID, deployment.Name)
	//}
}
