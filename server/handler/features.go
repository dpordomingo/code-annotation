package handler

import (
	"net/http"

	"github.com/src-d/code-annotation/server/model"
	"github.com/src-d/code-annotation/server/repository"
	"github.com/src-d/code-annotation/server/serializer"
)

// GetFeatures returns a function that returns a *serializer.Response
// with the list of features for the pairId from the Request payload
func GetFeatures(filePairRepo *repository.FilePairs) RequestProcessFunc {
	return func(r *http.Request) (*serializer.Response, error) {
		filePairID, err := urlParamInt(r, "pairId")
		if err != nil {
			return nil, err
		}

		filePair, err := filePairRepo.GetByID(filePairID)
		if err != nil {
			return nil, err
		}

		featuresA, featuresB, score, err := getFeatures(filePair)
		if err != nil {
			return nil, err
		}

		return serializer.NewFeaturesResponse(featuresA, featuresB, score), nil
	}
}

// TODO (dpordomingo): in the future it should take the UAST of both blobs DB
// and make a request to ML feature extractor API
func getFeatures(pair *model.FilePair) ([]*model.Feature, []*model.Feature, *model.Feature, error) {
	//blobIDA := pair.Left.BlobID
	//blobIDB := pair.Right.BlobID

	featuresA := []*model.Feature{}
	featuresB := []*model.Feature{}

	score := model.Feature{Name: "score", Weight: pair.Score}

	return featuresA, featuresB, &score, nil
}
