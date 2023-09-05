package helpers

import (
	"encoding/json"
	"facturaexpress/models"
)

func UnmarshalServices(data []byte) ([]models.Service, error) {
	var services []models.Service
	err := json.Unmarshal(data, &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}
