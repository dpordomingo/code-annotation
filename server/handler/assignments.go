package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/src-d/code-annotation/server/repository"
	"github.com/src-d/code-annotation/server/serializer"
	"github.com/src-d/code-annotation/server/service"
)

// GetAssignmentsForUserExperiment returns a function that returns a *serializer.Response
// with the assignments for the logged user and a passed experiment
// if these assignments do not already exist, they are created in advance
func GetAssignmentsForUserExperiment(repo *repository.Assignments) RequestProcessFunc {
	return func(r *http.Request) (*serializer.Response, error) {
		experimentID, err := urlParamInt(r, "experimentId")
		if err != nil {
			return nil, err
		}

		userID, err := service.GetUserID(r.Context())
		if err != nil {
			return nil, err
		}

		assignments, err := repo.GetAll(userID, experimentID)
		if err == repository.ErrNoAssignmentsInitialized {
			if assignments, err = repo.Initialize(userID, experimentID); err != nil {
				return nil, fmt.Errorf("no available assignments")
			}
		}

		return serializer.NewAssignmentsResponse(assignments), nil
	}
}

type assignmentRequest struct {
	Answer   string `json:"answer"`
	Duration int    `json:"duration"`
}

// SaveAssignment returns a function that saves the user answers as passed in the body request
func SaveAssignment(repo *repository.Assignments) RequestProcessFunc {
	return func(r *http.Request) (*serializer.Response, error) {
		pairID, err := urlParamInt(r, "pairId")
		if err != nil {
			return nil, err
		}

		userID, err := service.GetUserID(r.Context())
		if err != nil {
			return nil, err
		}

		var assignmentRequest assignmentRequest
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err == nil {
			err = json.Unmarshal(body, &assignmentRequest)
		}

		if err != nil {
			return nil, fmt.Errorf("payload could not be read")
		}

		err = repo.Update(userID, pairID, assignmentRequest.Answer, assignmentRequest.Duration)
		if err != nil {
			return nil, fmt.Errorf("answer could not be saved")
		}

		return serializer.NewCountResponse(1), nil
	}
}
