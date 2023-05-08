package page_bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (cf *CFPagesUploader) GetAllDeployments() ([]Deployment, error) {
	// get all deployments (GET https://api.cloudflare.com/client/v4/accounts/887ebfb71fdc5f2a59d20ab48eb4c9b2/pages/projects/kite-cms/deployments)

	// Create client
	client := &http.Client{}

	// Create request
	reqURL := fmt.Sprintf("%s/accounts/%s/pages/projects/%s/deployments", cf.BaseURL, cf.AccountID, cf.ProjectName)
	req, err := http.NewRequest("GET", reqURL, nil)

	// Headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cf.APIKey))

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := io.ReadAll(resp.Body)

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))

	var deploymentsResp DeploymentsResponse
	err = json.Unmarshal(respBody, &deploymentsResp)
	if err != nil {
		return nil, err
	}
	fmt.Printf("deploymentsResp: %+v\n", deploymentsResp)
	return deploymentsResp.Result, nil
}

func (cf *CFPagesUploader) DeleteDeployment(deploymentID string) error {
	// delete deployment (DELETE https://api.cloudflare.com/client/v4/accounts/887ebfb71fdc5f2a59d20ab48eb4c9b2/pages/projects/kite-cms/deployments/1dd9af2c-7c51-4a32-b2cf-b15c9f2728d1)

	// Create client
	client := &http.Client{}

	// Create request
	reqURL := fmt.Sprintf("%s/accounts/%s/pages/projects/%s/deployments/%s", cf.BaseURL, cf.AccountID, cf.ProjectName, deploymentID)
	req, err := http.NewRequest("DELETE", reqURL, nil)

	// Headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cf.APIKey))

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := io.ReadAll(resp.Body)

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))
	// Handle errors
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete deployment: status %s, body %s", resp.Status, string(respBody))
	}

	return nil
}

type Deployment struct {
	ID         string    `json:"id"`
	ModifiedOn time.Time `json:"modified_on"`
}

type DeploymentsResponse struct {
	Result []Deployment `json:"result"`
}

func (cf *CFPagesUploader) DeleteAllButLatestDeployment() {
	deployments, err := cf.GetAllDeployments()
	if err != nil {
		fmt.Println("Error getting all deployments:", err)
		return
	}

	if len(deployments) <= 1 {
		return
	}

	latest := deployments[0]
	for _, deployment := range deployments {
		if deployment.ModifiedOn.After(latest.ModifiedOn) {
			latest = deployment
		}
	}

	for _, deployment := range deployments {
		if deployment.ID != latest.ID {
			err := cf.DeleteDeployment(deployment.ID)
			if err != nil {
				fmt.Println("Error deleting deployment:", err)
			} else {
				fmt.Printf("Deleted deployment %s\n", deployment.ID)
			}
		}
	}
}

func sendGetAllDeployments() (*http.Response, error) {
	// get all deployments (GET https://api.cloudflare.com/client/v4/accounts/887ebfb71fdc5f2a59d20ab48eb4c9b2/pages/projects/kite-cms/deployments)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/accounts/887ebfb71fdc5f2a59d20ab48eb4c9b2/pages/projects/kite-cms/deployments", nil)

	// Headers
	req.Header.Add("Authorization", "Bearer JLM65BNN3hr-kCn2IIn8l3a3_aD69DZlWqkZP-gn")
	req.Header.Add("Cookie", "__cflb=0H28vgHxwvgAQtjUGUFqYFDiSDreGJnUkHa3tUqP6Lu; __cfruid=97df13ffd48ac88340fd2be3699cc69319a984b3-1683567202")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	return resp, nil
}
