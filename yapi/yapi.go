package yapi

import (
	"encoding/json"
	"github.com/imroc/req/v3"
	"github.com/nyg123/protoc-gen-openapiv2-yapi/yapi/def"
	"strings"
	"time"
)

var client *req.Client

func Run(data []byte, url string, token string) error {
	client = req.C().SetTimeout(20 * time.Second)
	var swagger map[string]interface{}
	globalTitle := map[string]string{}
	err := json.Unmarshal(data, &swagger)
	if err != nil {
		return err
	}
	hearders := []def.Header{}
	if header, ok := swagger["x-header"]; ok {
		hearders = getHeader(header)
		swagger["x-header"] = hearders
	}
	name := ""
	if tags, ok := swagger["tags"]; ok {
		tagsArr := tags.([]interface{})
		tmp := tagsArr[0]
		tmpMap := tmp.(map[string]interface{})
		tmpMap["name"] = tmpMap["description"].(string)
		tagsArr[0] = tmpMap
		swagger["tags"] = tagsArr
		name = tmpMap["name"].(string)
	}
	if definitions, ok := swagger["definitions"]; ok {
		definitionsMap := definitions.(map[string]interface{})
		for _, v := range definitionsMap {
			vMap := v.(map[string]interface{})
			if title, ok := vMap["title"]; ok {
				vMap["description"] = title.(string)
			}
			if properties, ok := vMap["properties"]; ok {
				vMap["properties"] = handleProperties(properties, &globalTitle)
			}
		}
		swagger["definitions"] = definitionsMap
	}
	if definitions, ok := swagger["definitions"]; ok {
		definitionsMap := definitions.(map[string]interface{})
		for k, v := range definitionsMap {
			vMap := v.(map[string]interface{})
			if t, ok := vMap["type"]; ok {
				if t.(string) == "object" {
					vMap["description"] = globalTitle[k]
				}
			}
		}
		swagger["definitions"] = definitionsMap
	}

	if paths, ok := swagger["paths"]; ok {
		pathsMap := paths.(map[string]interface{})
		for k1, v := range pathsMap {
			vMap := v.(map[string]interface{})
			for k2, v2 := range vMap {
				v2Map := v2.(map[string]interface{})
				v2Map["tags"] = []string{name}
				parameters := []interface{}{}
				if parametersTmp, ok := v2Map["parameters"]; ok {
					parameters = parametersTmp.([]interface{})
				}
				for _, header := range hearders {
					parameters = append(parameters, header)
				}
				if header, ok := v2Map["x-header"]; ok {
					hearders = getHeader(header)
					for _, header := range hearders {
						parameters = append(parameters, header)
					}
				}
				v2Map["parameters"] = parameters
				vMap[k2] = v2Map
			}
			pathsMap[k1] = vMap
		}
	}
	// 写入文件
	data, err = json.MarshalIndent(swagger, "", "    ")
	if err != nil {
		return err
	}
	// POST 请求
	return PostYapi(data, url, token)
}

func PostYapi(d []byte, url string, token string) error {
	data := map[string]string{
		"type":  "swagger",
		"json":  string(d),
		"merge": "merge",
		"token": token,
	}
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = client.R().SetHeader("Content-Type", "application/json").
		SetBody(jsonStr).Post(url + "/api/open/import_data")
	if err != nil {
		return err
	}
	return nil
}

func getHeader(header interface{}) []def.Header {
	h := header.(map[string]interface{})
	var headers []def.Header
	for k, v := range h {
		t := v.(map[string]interface{})
		required := false
		if t["required"] != nil {
			required = t["required"].(bool)
		}
		headers = append(headers, def.Header{
			Name:        k,
			Description: t["description"].(string),
			Type:        "string",
			In:          "header",
			Required:    required,
		})
	}
	return headers
}

func handleProperties(properties interface{}, globalTitle *map[string]string) map[string]interface{} {
	propertiesMap := properties.(map[string]interface{})
	for _, v := range propertiesMap {
		vMap := v.(map[string]interface{})
		if title, ok := vMap["title"]; ok {
			if description, ok := vMap["description"]; ok {
				vMap["description"] = description.(string) + title.(string)
			} else {
				vMap["description"] = title.(string)
			}
			if ref, ok := vMap["$ref"]; ok {
				refStr := strings.Split(ref.(string), "/")[2]
				(*globalTitle)[refStr] = vMap["description"].(string)
			}
		}
		if properties, ok := vMap["properties"]; ok {
			vMap["properties"] = handleProperties(properties, globalTitle)
		}
	}
	return propertiesMap
}
