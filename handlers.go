package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ProxyService struct {
	cfg    *config
	client *http.Client
}

func NewProxyService(cfg *config) (*ProxyService, error) {
	client, err := getClient(cfg)
	if nil != err {
		return nil, err
	}

	return &ProxyService{
		cfg:    cfg,
		client: client,
	}, nil
}

type Pong struct {
	Now    int    `json:"now"`
	Status string `json:"status"`
	Ns1    string `json:"ns1"`
}

func (s *ProxyService) pong(c *gin.Context) {
	// log.Println("Entering pong function")
	// log.Println("Received ping request")
	c.JSON(http.StatusOK, Pong{
		Now:    time.Now().Second(),
		Status: "ok",
		Ns1:    "200 OK",
	})
	// log.Println("Sent pong response")
	// log.Println("Exiting pong function")
}

func (s *ProxyService) models(c *gin.Context) {
	// log.Println("Entering models function")
	// log.Println("Received models request")
	c.JSON(http.StatusOK, gin.H{
		"data": []gin.H{
			{
				"capabilities": gin.H{
					"family": "gpt-3.5-turbo",
					"object": "model_capabilities",
					"type":   "chat",
				},
				"id":      "gpt-3.5-turbo",
				"name":    "GPT 3.5 Turbo",
				"object":  "model",
				"version": "gpt-3.5-turbo-0613",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-3.5-turbo",
					"object": "model_capabilities",
					"type":   "chat",
				},
				"id":      "gpt-3.5-turbo-0613",
				"name":    "GPT 3.5 Turbo (2023-06-13)",
				"object":  "model",
				"version": "gpt-3.5-turbo-0613",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4",
					"object": "model_capabilities",
					"type":   "chat",
				},
				"id":      "gpt-4",
				"name":    "GPT 4",
				"object":  "model",
				"version": "gpt-4-0613",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4",
					"object": "model_capabilities",
					"type":   "chat",
				},
				"id":      "gpt-4-0613",
				"name":    "GPT 4 (2023-06-13)",
				"object":  "model",
				"version": "gpt-4-0613",
			},
			{
				"capabilities": gin.H{
					"family": "gpt-4-turbo",
					"object": "model_capabilities",
					"type":   "chat",
				},
				"id":      "gpt-4-0125-preview",
				"name":    "GPT 4 Turbo (2024-01-25 Preview)",
				"object":  "model",
				"version": "gpt-4-0125-preview",
			},
			{
				"capabilities": gin.H{
					"family": "text-embedding-ada-002",
					"object": "model_capabilities",
					"type":   "embeddings",
				},
				"id":      "text-embedding-ada-002",
				"name":    "Embedding V2 Ada",
				"object":  "model",
				"version": "text-embedding-ada-002",
			},
			{
				"capabilities": gin.H{
					"family": "text-embedding-ada-002",
					"object": "model_capabilities",
					"type":   "embeddings",
				},
				"id":      "text-embedding-ada-002-index",
				"name":    "Embedding V2 Ada (Index)",
				"object":  "model",
				"version": "text-embedding-ada-002",
			},
			{
				"capabilities": gin.H{
					"family": "text-embedding-3-small",
					"object": "model_capabilities",
					"type":   "embeddings",
				},
				"id":      "text-embedding-3-small",
				"name":    "Embedding V3 small",
				"object":  "model",
				"version": "text-embedding-3-small",
			},
			{
				"capabilities": gin.H{
					"family": "text-embedding-3-small",
					"object": "model_capabilities",
					"type":   "embeddings",
				},
				"id":      "text-embedding-3-small-inference",
				"name":    "Embedding V3 small (Inference)",
				"object":  "model",
				"version": "text-embedding-3-small",
			},
		},
		"object": "list",
	})
	// log.Println("Sent models response")
	// log.Println("Exiting models function")
}

func (s *ProxyService) completions(c *gin.Context) {
	// log.Println("Entering completions function")
	ctx := c.Request.Context()

	// log.Println("Received completions request")
	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		log.Println("Failed to read request body:", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	// log.Printf("Request body: %s\n", string(body))

	model := gjson.GetBytes(body, "model").String()
	if mapped, ok := s.cfg.ChatModelMap[model]; ok {
		model = mapped
	} else {
		model = s.cfg.ChatModelDefault
	}
	body, _ = sjson.SetBytes(body, "model", model)

	if !gjson.GetBytes(body, "function_call").Exists() {
		messages := gjson.GetBytes(body, "messages").Array()
		lastIndex := len(messages) - 1
		if !strings.Contains(messages[lastIndex].Get("content").String(), "Respond in the following locale") {
			locale := s.cfg.ChatLocale
			if locale == "" {
				locale = "zh_CN"
			}
			body, _ = sjson.SetBytes(body, "messages."+strconv.Itoa(lastIndex)+".content", messages[lastIndex].Get("content").String()+"Respond in the following locale: "+locale+".")
		}
	}

	body, _ = sjson.DeleteBytes(body, "intent")
	body, _ = sjson.DeleteBytes(body, "intent_threshold")
	body, _ = sjson.DeleteBytes(body, "intent_content")

	if int(gjson.GetBytes(body, "max_tokens").Int()) > s.cfg.ChatMaxTokens {
		body, _ = sjson.SetBytes(body, "max_tokens", s.cfg.ChatMaxTokens)
	}

	proxyUrl := s.cfg.ChatApiBase + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyUrl, io.NopCloser(bytes.NewBuffer(body)))
	if nil != err {
		log.Println("Failed to create request:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.ChatApiKey)
	if "" != s.cfg.ChatApiOrganization {
		req.Header.Set("OpenAI-Organization", s.cfg.ChatApiOrganization)
	}
	if "" != s.cfg.ChatApiProject {
		req.Header.Set("OpenAI-Project", s.cfg.ChatApiProject)
	}

	resp, err := s.client.Do(req)
	if nil != err {
		if errors.Is(err, context.Canceled) {
			log.Println("Request canceled:", err)
			c.AbortWithStatus(http.StatusRequestTimeout)
			return
		}

		log.Println("Request conversation failed:", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer closeIO(resp.Body)

	if resp.StatusCode != http.StatusOK { // log
		body, _ := io.ReadAll(resp.Body)
		log.Println("Request completions failed:", string(body))

		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	c.Status(resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")
	if "" != contentType {
		c.Header("Content-Type", contentType)
	}

	_, _ = io.Copy(c.Writer, resp.Body)
	// log.Println("Sent completions response")
	// log.Println("Exiting completions function")
}

func (s *ProxyService) codeCompletions(c *gin.Context) {
	// log.Println("Entering codeCompletions function")
	ctx := c.Request.Context()

	// log.Println("Received code completions request")
	time.Sleep(200 * time.Millisecond)
	if ctx.Err() != nil {
		log.Println("Request timeout:", ctx.Err())
		abortCodex(c, http.StatusRequestTimeout)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		log.Println("Failed to read request body:", err)
		abortCodex(c, http.StatusBadRequest)
		return
	}
	// log.Printf("Request body: %s\n", string(body))

	body = ConstructRequestBody(body, s.cfg)

	proxyUrl := s.cfg.CodexApiBase + "/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyUrl, io.NopCloser(bytes.NewBuffer(body)))
	if nil != err {
		log.Println("Failed to create request:", err)
		abortCodex(c, http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.CodexApiKey)
	if "" != s.cfg.CodexApiOrganization {
		req.Header.Set("OpenAI-Organization", s.cfg.CodexApiOrganization)
	}
	if "" != s.cfg.CodexApiProject {
		req.Header.Set("OpenAI-Project", s.cfg.CodexApiProject)
	}

	resp, err := s.client.Do(req)
	if nil != err {
		if errors.Is(err, context.Canceled) {
			log.Println("Request canceled:", err)
			abortCodex(c, http.StatusRequestTimeout)
			return
		}

		log.Println("Request completions failed:", err.Error())
		abortCodex(c, http.StatusInternalServerError)
		return
	}
	defer closeIO(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("Request completions failed:", string(body))

		abortCodex(c, resp.StatusCode)
		return
	}

	c.Status(resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")
	if "" != contentType {
		c.Header("Content-Type", contentType)
	}

	_, _ = io.Copy(c.Writer, resp.Body)
	// log.Println("Sent code completions response")
	// log.Println("Exiting codeCompletions function")
}
