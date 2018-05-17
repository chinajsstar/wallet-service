package jsonparse

import (
	"encoding/json"
	"math"
	"strings"
)

type JsonParse struct {
	jsonValue interface{}
}

func Parse(data string) (*JsonParse, error) {
	p := new(JsonParse)
	err := json.Unmarshal([]byte(data), &p.jsonValue)
	return p, err
}

func (p *JsonParse) UserKey() (string, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["user_key"]; ok {
			if value, ok := value.(string); ok {
				return value, true
			}
		}
	}
	return "", false
}

func (p *JsonParse) UserOrderID() (string, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["user_order_id"]; ok {
			if value, ok := value.(string); ok {
				return value, true
			}
		}
	}
	return "", false
}

func (p *JsonParse) AssetName() (string, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["asset_name"]; ok {
			if value, ok := value.(string); ok {
				return strings.ToLower(value), true
			}
		}
	}
	return "", false
}

func (p *JsonParse) AssetNameArray() ([]string, bool) {
	arr := make([]string, 0)
	if value, ok := p.jsonValue.([]interface{}); ok {
		for _, value := range value {
			if value, ok := value.(string); ok {
				arr = append(arr, strings.ToLower(value))
			}
		}
	}
	return arr, len(arr) > 0
}

func (p *JsonParse) Address() (string, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["address"]; ok {
			if value, ok := value.(string); ok {
				return value, true
			}
		}
	}
	return "", false
}

func (p *JsonParse) Count() (int, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["count"]; ok {
			if value, ok := value.(float64); ok {
				return int(value), true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) Amount() (int64, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["amount"]; ok {
			if value, ok := value.(float64); ok {
				return int64(value * math.Pow10(8)), true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) TransType() (int, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["trans_type"]; ok {
			if value, ok := value.(float64); ok {
				return int(value), true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) Status() (int, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["status"]; ok {
			if value, ok := value.(float64); ok {
				return int(value), true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) MaxAmount() (float64, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["max_amount"]; ok {
			if value, ok := value.(float64); ok {
				return value, true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) MinAmount() (float64, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["min_amount"]; ok {
			if value, ok := value.(float64); ok {
				return value, true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) MaxMessageID() (float64, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["max_msg_id"]; ok {
			if value, ok := value.(float64); ok {
				return value, true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) MinMessageID() (float64, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["min_msg_id"]; ok {
			if value, ok := value.(float64); ok {
				return value, true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) MaxCreateTime() (int64, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["max_create_time"]; ok {
			if value, ok := value.(float64); ok {
				return int64(value), true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) MinCreateTime() (int64, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["min_create_time"]; ok {
			if value, ok := value.(float64); ok {
				return int64(value), true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) MaxUpdateTime() (int64, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["max_update_time"]; ok {
			if value, ok := value.(float64); ok {
				return int64(value), true
			}
		}
	}
	return 0, false
}

func (p *JsonParse) MinUpdateTime() (int64, bool) {
	if value, ok := p.jsonValue.(map[string]interface{}); ok {
		if value, ok := value["min_update_time"]; ok {
			if value, ok := value.(float64); ok {
				return int64(value), true
			}
		}
	}
	return 0, false
}
