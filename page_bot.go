package page_bot

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CFPagesUploader struct {
	AccountID   string
	ProjectName string
	APIKey      string
	BaseURL     string
	CachedJWT   string
	ExpiresAt   time.Time
}

type FileContent struct {
	Filename    string
	Content     []byte
	ContentType string
}

type PageData struct {
	Title   string
	Content string
}

func NewCFPagesUploader(accountID, projectName, apiKey, baseURL string) *CFPagesUploader {
	return &CFPagesUploader{
		AccountID:   accountID,
		ProjectName: projectName,
		APIKey:      apiKey,
		BaseURL:     baseURL,
	}
}

func (c *CFPagesUploader) getJWT() (string, error) {
	// 如果缓存的 JWT 未过期，则返回缓存的 JWT
	now := time.Now()
	if c.CachedJWT != "" && c.ExpiresAt.After(now) {
		return c.CachedJWT, nil
	}

	// 其他情况下，发送请求以获取新的 JWT
	url := fmt.Sprintf("%s/accounts/%s/pages/projects/%s/upload-token", c.BaseURL, c.AccountID, c.ProjectName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Set("User-Agent", "CFPagesUploader-Go/1.0.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Success bool
		Result  struct {
			JWT string `json:"jwt"`
		}
		Errors []interface{}
	}

	err = json.Unmarshal(body, &result)

	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("Failed to get JWT: %v", result.Errors)
	}

	// 在缓存中存储新的 JWT 和过期时间
	c.CachedJWT = result.Result.JWT
	c.ExpiresAt = now.Add(4 * time.Minute) // 将过期时间设置为 4 分钟，以确保在实际到期之前使用新的 JWT

	return c.CachedJWT, nil
}

func computeHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

func createManifest(indexHTMLContent string) (string, error) {
	hash := computeHash(indexHTMLContent)
	manifest := map[string]string{
		"/index.html": hash,
	}

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	return string(manifestJSON), nil
}

func (uploader *CFPagesUploader) uploadAsset(hash, content, contentType string) (string, error) {
	fmt.Printf("ready to upload %s", hash)
	jwt, err := uploader.getJWT()
	url := fmt.Sprintf("%s/pages/assets/upload", uploader.BaseURL)
	// 创建上传请求的 JSON 数据
	data := []map[string]interface{}{
		{
			"key":   hash,
			"value": content,
			"metadata": map[string]string{
				"contentType": contentType,
			},
			"base64": true,
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	body := strings.NewReader(string(jsonData))

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload asset, status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return "", err
	}

	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &jsonResponse); err != nil {
		return "", err
	}

	if jsonResponse["success"].(bool) {
		fmt.Printf("upload response: %s\n", jsonResponse)
		return hash, nil
	} else {
		return "", fmt.Errorf("upload not successful")
	}
}

func (uploader *CFPagesUploader) upsertHashes(hashes []string) error {
	jwt, err := uploader.getJWT()
	if err != nil {
		panic(err)
	}
	//fmt.Printf("jwt token is-- %s --", jwt)

	body, err := json.Marshal(map[string]interface{}{
		"hashes": hashes,
	})

	url := fmt.Sprintf("%s/pages/assets/upsert-hashes", uploader.BaseURL)

	//hashesMaps := make([]map[string]string, len(hashes))
	//for i, hash := range hashes {
	//	hashesMaps[i] = map[string]string{
	//		"hash": hash,
	//	}
	//}
	//
	//// 将哈希值放在名为 "hashes" 的属性中
	//data := map[string]interface{}{
	//	"hashes": hashesMaps,
	//}

	//hashesJSON, err := json.Marshal(data)
	//if err != nil {
	//	return err
	//}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upsert hashes, status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &jsonResponse); err != nil {
		return err
	}

	if jsonResponse["success"] != true {
		return fmt.Errorf("failed to upsert hashes: %v", jsonResponse["errors"])
	}

	return nil
}

func generateMD5Hash(input string) string {
	hasher := md5.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

func (uploader *CFPagesUploader) DeployFiles(generatedHTMLFiles []FileContent) error {
	uploadedHashes := make([]string, len(generatedHTMLFiles))
	hashChan := make(chan struct {
		hash  string
		index int
	}, len(generatedHTMLFiles))
	errChan := make(chan error, len(generatedHTMLFiles))

	// 计算所有文件的哈希值
	fileHashes := make([]string, len(generatedHTMLFiles))
	for i, file := range generatedHTMLFiles {
		fileHashes[i] = generateMD5Hash(string(file.Content))
	}

	// 获取服务器上尚不存在的哈希值
	missingHashes, err := uploader.getMissingHashes(fileHashes)
	if err != nil {
		return fmt.Errorf("failed to get missing hashes: %w", err)
	}
	fmt.Printf("missingHashes %s", missingHashes)
	missingHashSet := make(map[string]struct{}, len(missingHashes))
	for _, hash := range missingHashes {
		missingHashSet[hash] = struct{}{}
	}

	var wg sync.WaitGroup
	for i, file := range generatedHTMLFiles {
		if _, exists := missingHashSet[fileHashes[i]]; exists {
			wg.Add(1)
			go func(index int, fileContent FileContent) {
				defer wg.Done()

				fileHash := fileHashes[index]
				base64Content := base64.StdEncoding.EncodeToString(fileContent.Content)

				uploadedHash, err := uploader.uploadAsset(fileHash, base64Content, fileContent.ContentType)
				if err != nil {
					errChan <- fmt.Errorf("failed to upload file %s: %w", fileContent.Filename, err)
					return
				}

				hashChan <- struct {
					hash  string
					index int
				}{uploadedHash, index}
			}(i, file)
		} else {
			uploadedHashes[i] = fileHashes[i]
		}
	}

	go func() {
		wg.Wait()
		close(hashChan)
		close(errChan)
	}()

	for range missingHashes {
		select {
		case hashData := <-hashChan:
			uploadedHashes[hashData.index] = hashData.hash
		case err := <-errChan:
			return err
		}
	}

	fmt.Printf("wait for upsert hashes %s", uploadedHashes)
	// 更新hash
	err = uploader.upsertHashes(uploadedHashes)
	if err != nil {
		log.Fatalf("Failed to upsert hashes: %v", err)
	}

	manifest := make(map[string]string, len(generatedHTMLFiles))
	for i, file := range generatedHTMLFiles {
		manifest[file.Filename] = uploadedHashes[i]
	}

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	fmt.Println("manifestJSON:", string(manifestJSON))
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	manifestFormField, err := writer.CreateFormField("manifest")
	if err != nil {
		log.Fatal(err)
		return err
	}

	_, err = manifestFormField.Write([]byte(manifestJSON))
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = writer.Close()
	if err != nil {
		println(err)
		log.Fatal(err)
		return err
	}

	// 发送创建部署请求
	url := fmt.Sprintf("%s/accounts/%s/pages/projects/%s/deployments", uploader.BaseURL, uploader.AccountID, uploader.ProjectName)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		println(err)
		log.Fatal(err)
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", uploader.APIKey))
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		println(err)
		log.Fatal(err)
		return err
	}
	defer resp.Body.Close()
	fmt.Printf("deploy response: %v", resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to deploy, status code: %d", resp.StatusCode)
	}

	return nil
}

type GetMissingHashesResponse struct {
	Success bool     `json:"success"`
	Result  []string `json:"result"`
	Errors  []string `json:"errors"`
}

func (uploader *CFPagesUploader) getMissingHashes(hashes []string) ([]string, error) {
	jwt, err := uploader.getJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT: %w", err)
	}
	url := fmt.Sprintf("%s/pages/assets/check-missing", uploader.BaseURL)
	body, err := json.Marshal(map[string]interface{}{
		"hashes": hashes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var res GetMissingHashesResponse
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if res.Success {
		fmt.Printf("sucess result: %s", res.Result)
		return res.Result, nil
	}

	return nil, fmt.Errorf("failed to get missing hashes: %v", res.Errors)
}
