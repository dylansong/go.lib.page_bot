package page_bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

func (uploader *CFPagesUploader) GetAllDeployments() []byte {
	// get all deployments (GET https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments)

	// Create client
	client := &http.Client{}

	// Create request
	url := fmt.Sprintf("%s/accounts/%s/pages/projects/%s/deployments", uploader.BaseURL, uploader.AccountID, uploader.ProjectName)
	req, err := http.NewRequest("GET", url, nil)

	// Headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", uploader.APIKey))

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := io.ReadAll(resp.Body)

	return respBody
}

func (uploader *CFPagesUploader) DeleteDeployment(deploymentID string) {
	// delete deployment (DELETE https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments/{deployment_id})

	// Create client
	client := &http.Client{}

	// Create request
	url := fmt.Sprintf("%s/accounts/%s/pages/projects/%s/deployments/%s", uploader.BaseURL, uploader.AccountID, uploader.ProjectName, deploymentID)
	req, err := http.NewRequest("DELETE", url, nil)

	// Headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", uploader.APIKey))

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
}

type Deployment struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	State      string    `json:"state"`
	CreatedOn  time.Time `json:"created_on"`
	ModifiedOn time.Time `json:"modified_on"`
}

type Deployments []Deployment

func (slice Deployments) Len() int {
	return len(slice)
}

func (slice Deployments) Less(i, j int) bool {
	return slice[i].ModifiedOn.Before(slice[j].ModifiedOn)
}

func (slice Deployments) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (uploader *CFPagesUploader) DeleteAllDeploymentsExceptLatest() {
	// Get all deployments
	respBody := uploader.GetAllDeployments()
	var deployments Deployments
	json.Unmarshal(respBody, &deployments)

	// Sort deployments by modified_on
	sort.Sort(deployments)

	// Delete all deployments except the latest one
	latestDeployment := deployments[len(deployments)-1]
	for _, deployment := range deployments[:len(deployments)-1] {
		uploader.DeleteDeployment(deployment.ID)
	}

	fmt.Printf("Deleted %d deployments. Latest deployment: %s\n", len(deployments)-1, latestDeployment.ID)
}
