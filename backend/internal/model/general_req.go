package model

import "dd-prediction-api/pkg/repository"

type OptsReq struct {
	Opts repository.Options `json:"opts" query:"opts"`
}
