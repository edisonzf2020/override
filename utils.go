package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"strings"
)

const StableCodeModelPrefix = "stable-code"

const DeepSeekCoderModel = "deepseek-coder"

func abortCodex(c *gin.Context, status int) {
	c.Header("Content-Type", "text/event-stream")

	c.String(status, "data: [DONE]\n")
	c.Abort()
}

func closeIO(c io.Closer) {
	err := c.Close()
	if nil != err {
		log.Println(err)
	}
}

func ConstructRequestBody(body []byte, cfg *config) []byte {
	body, _ = sjson.DeleteBytes(body, "extra")
	body, _ = sjson.DeleteBytes(body, "nwo")
	body, _ = sjson.SetBytes(body, "model", cfg.CodeInstructModel)

	if int(gjson.GetBytes(body, "max_tokens").Int()) > cfg.CodexMaxTokens {
		body, _ = sjson.SetBytes(body, "max_tokens", cfg.CodexMaxTokens)
	}

	if strings.Contains(cfg.CodeInstructModel, StableCodeModelPrefix) {
		return constructWithStableCodeModel(body)
	} else if strings.HasPrefix(cfg.CodeInstructModel, DeepSeekCoderModel) {
		if gjson.GetBytes(body, "n").Int() > 1 {
			body, _ = sjson.SetBytes(body, "n", 1)
		}
	}

	if strings.HasSuffix(cfg.ChatApiBase, "chat") {
		// @Todo  constructWithChatModel
		// 如果code base以chat结尾则构建chatModel，暂时没有好的prompt
	}

	return body
}

func constructWithStableCodeModel(body []byte) []byte {
	suffix := gjson.GetBytes(body, "suffix")
	prompt := gjson.GetBytes(body, "prompt")
	content := fmt.Sprintf("<fim_prefix>%s<fim_suffix>%s<fim_middle>", prompt, suffix)

	// 创建新的 JSON 对象并添加到 body 中
	messages := []map[string]string{
		{
			"role":    "user",
			"content": content,
		},
	}
	return constructWithChatModel(body, messages)
}

func constructWithChatModel(body []byte, messages interface{}) []byte {

	body, _ = sjson.SetBytes(body, "messages", messages)

	// fmt.Printf("Request Body: %s\n", body)
	// 2. 将转义的字符替换回原来的字符
	jsonStr := string(body)
	jsonStr = strings.ReplaceAll(jsonStr, "\\u003c", "<")
	jsonStr = strings.ReplaceAll(jsonStr, "\\u003e", ">")
	return []byte(jsonStr)
}
