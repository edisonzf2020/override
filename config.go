package main

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const DefaultInstructModel = "gpt-3.5-turbo-instruct"

type config struct {
	Bind                 string            `json:"bind"`
	ProxyUrl             string            `json:"proxy_url"`
	Timeout              int               `json:"timeout"`
	CodexApiBase         string            `json:"codex_api_base"`
	CodexApiKey          string            `json:"codex_api_key"`
	CodexApiOrganization string            `json:"codex_api_organization"`
	CodexApiProject      string            `json:"codex_api_project"`
	CodexMaxTokens       int               `json:"codex_max_tokens"`
	CodeInstructModel    string            `json:"code_instruct_model"`
	ChatApiBase          string            `json:"chat_api_base"`
	ChatApiKey           string            `json:"chat_api_key"`
	ChatApiOrganization  string            `json:"chat_api_organization"`
	ChatApiProject       string            `json:"chat_api_project"`
	ChatMaxTokens        int               `json:"chat_max_tokens"`
	ChatModelDefault     string            `json:"chat_model_default"`
	ChatModelMap         map[string]string `json:"chat_model_map"`
	ChatLocale           string            `json:"chat_locale"`
	AuthToken            string            `json:"auth_token"`
}

func readConfig() *config {
	content, err := os.ReadFile("config.json")
	if nil != err {
		log.Fatal(err)
	}

	_cfg := &config{}
	err = json.Unmarshal(content, &_cfg)
	if nil != err {
		log.Fatal(err)
	}

	v := reflect.ValueOf(_cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("json")
		if tag == "" {
			continue
		}

		value, exists := os.LookupEnv("OVERRIDE_" + strings.ToUpper(tag))
		if !exists {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(value)
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(value); err == nil {
				field.SetBool(boolValue)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				field.SetInt(intValue)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if uintValue, err := strconv.ParseUint(value, 10, 64); err == nil {
				field.SetUint(uintValue)
			}
		case reflect.Float32, reflect.Float64:
			if floatValue, err := strconv.ParseFloat(value, field.Type().Bits()); err == nil {
				field.SetFloat(floatValue)
			}
		}
	}
	if _cfg.CodeInstructModel == "" {
		_cfg.CodeInstructModel = DefaultInstructModel
	}

	if _cfg.CodexMaxTokens == 0 {
		_cfg.CodexMaxTokens = 500
	}

	if _cfg.ChatMaxTokens == 0 {
		_cfg.ChatMaxTokens = 4096
	}

	return _cfg
}
