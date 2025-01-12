package storage

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/ssi-sdk/credential/manifest"

	cred "github.com/tbd54566975/ssi-service/internal/credential"
	"github.com/tbd54566975/ssi-service/internal/keyaccess"
	"github.com/tbd54566975/ssi-service/internal/util"
	"github.com/tbd54566975/ssi-service/pkg/storage"
)

const (
	manifestNamespace    = "manifest"
	applicationNamespace = "application"
	responseNamespace    = "response"
)

type StoredManifest struct {
	ID          string                      `json:"id"`
	Issuer      string                      `json:"issuer"`
	Manifest    manifest.CredentialManifest `json:"manifest"`
	ManifestJWT keyaccess.JWT               `json:"manifestJwt"`
}

type StoredApplication struct {
	ID             string                         `json:"id"`
	ManifestID     string                         `json:"manifestId"`
	ApplicantDID   string                         `json:"applicantDid"`
	Application    manifest.CredentialApplication `json:"application"`
	Credentials    []cred.Container               `json:"credentials"`
	ApplicationJWT keyaccess.JWT                  `json:"applicationJwt"`
}

type StoredResponse struct {
	ID           string                      `json:"id"`
	ManifestID   string                      `json:"manifestId"`
	ApplicantDID string                      `json:"applicantId"`
	Response     manifest.CredentialResponse `json:"response"`
	Credentials  []cred.Container            `json:"credentials"`
	ResponseJWT  keyaccess.JWT               `json:"responseJwt"`
}

type Storage struct {
	db storage.ServiceStorage
}

func NewManifestStorage(db storage.ServiceStorage) (*Storage, error) {
	if db == nil {
		return nil, errors.New("bolt db reference is nil")
	}
	return &Storage{db: db}, nil
}

func (ms *Storage) StoreManifest(manifest StoredManifest) error {
	id := manifest.Manifest.ID
	if id == "" {
		err := errors.New("could not store manifest without an ID")
		logrus.WithError(err).Error()
		return err
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		errMsg := fmt.Sprintf("could not store manifest: %s", id)
		logrus.WithError(err).Error(errMsg)
		return errors.Wrapf(err, errMsg)
	}
	return ms.db.Write(manifestNamespace, id, manifestBytes)
}

func (ms *Storage) GetManifest(id string) (*StoredManifest, error) {
	manifestBytes, err := ms.db.Read(manifestNamespace, id)
	if err != nil {
		errMsg := fmt.Sprintf("could not get manifest: %s", id)
		logrus.WithError(err).Error(errMsg)
		return nil, errors.Wrapf(err, errMsg)
	}
	if len(manifestBytes) == 0 {
		err := fmt.Errorf("manifest not found with id: %s", id)
		logrus.WithError(err).Error("could not get manifest from storage")
		return nil, err
	}
	var stored StoredManifest
	if err := json.Unmarshal(manifestBytes, &stored); err != nil {
		errMsg := fmt.Sprintf("could not unmarshal stored manifest: %s", id)
		logrus.WithError(err).Error(errMsg)
		return nil, errors.Wrapf(err, errMsg)
	}
	return &stored, nil
}

// GetManifests attempts to get all stored manifests. It will return those it can even if it has trouble with some.
func (ms *Storage) GetManifests() ([]StoredManifest, error) {
	gotManifests, err := ms.db.ReadAll(manifestNamespace)
	if err != nil {
		errMsg := "could not get all manifests"
		logrus.WithError(err).Error(errMsg)
		return nil, errors.Wrap(err, errMsg)
	}
	if len(gotManifests) == 0 {
		logrus.Info("no manifests to get")
		return nil, nil
	}
	var stored []StoredManifest
	for _, manifestBytes := range gotManifests {
		var nextManifest StoredManifest
		if err = json.Unmarshal(manifestBytes, &nextManifest); err == nil {
			stored = append(stored, nextManifest)
		}
	}
	return stored, nil
}

func (ms *Storage) DeleteManifest(id string) error {
	if err := ms.db.Delete(manifestNamespace, id); err != nil {
		return util.LoggingErrorMsgf(err, "could not delete manifest: %s", id)
	}
	return nil
}

func (ms *Storage) StoreApplication(application StoredApplication) error {
	id := application.Application.ID
	if id == "" {
		return util.LoggingNewError("could not store application without an ID")
	}
	applicationBytes, err := json.Marshal(application)
	if err != nil {
		return util.LoggingErrorMsgf(err, "could not store application: %s", id)
	}
	return ms.db.Write(applicationNamespace, id, applicationBytes)
}

func (ms *Storage) GetApplication(id string) (*StoredApplication, error) {
	applicationBytes, err := ms.db.Read(applicationNamespace, id)
	if err != nil {
		return nil, util.LoggingErrorMsgf(err, "could not get application: %s", id)
	}
	if len(applicationBytes) == 0 {
		return nil, util.LoggingNewErrorf("could not get application from storage; application not found with id: %s", id)
	}
	var stored StoredApplication
	if err = json.Unmarshal(applicationBytes, &stored); err != nil {
		return nil, util.LoggingErrorMsgf(err, "could not unmarshal stored application: %s", id)
	}
	return &stored, nil
}

// GetApplications attempts to get all stored applications. It will return those it can even if it has trouble with some.
func (ms *Storage) GetApplications() ([]StoredApplication, error) {
	gotApplications, err := ms.db.ReadAll(applicationNamespace)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "could not get all applications")
	}
	if len(gotApplications) == 0 {
		logrus.Info("no applications to get")
		return nil, nil
	}
	var stored []StoredApplication
	for appKey, applicationBytes := range gotApplications {
		var nextApplication StoredApplication
		if err = json.Unmarshal(applicationBytes, &nextApplication); err == nil {
			stored = append(stored, nextApplication)
		} else {
			logrus.WithError(err).Errorf("could not unmarshal stored application: %s", appKey)
		}
	}
	return stored, nil
}

func (ms *Storage) DeleteApplication(id string) error {
	if err := ms.db.Delete(applicationNamespace, id); err != nil {
		return util.LoggingErrorMsgf(err, "could not delete application: %s", id)
	}
	return nil
}

func (ms *Storage) StoreResponse(response StoredResponse) error {
	id := response.Response.ID
	if id == "" {
		return util.LoggingNewError("could not store response without an ID")
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return util.LoggingErrorMsgf(err, "could not store response: %s", id)
	}
	return ms.db.Write(responseNamespace, id, responseBytes)
}

func (ms *Storage) GetResponse(id string) (*StoredResponse, error) {
	responseBytes, err := ms.db.Read(responseNamespace, id)
	if err != nil {
		return nil, util.LoggingErrorMsgf(err, "could not get response: %s", id)
	}
	if len(responseBytes) == 0 {
		return nil, util.LoggingErrorMsgf(err, "response not found with id: %s", id)
	}
	var stored StoredResponse
	if err = json.Unmarshal(responseBytes, &stored); err != nil {
		return nil, util.LoggingErrorMsgf(err, "could not unmarshal stored response: %s", id)
	}
	return &stored, nil
}

// GetResponses attempts to get all stored responses. It will return those it can even if it has trouble with some.
func (ms *Storage) GetResponses() ([]StoredResponse, error) {
	gotResponses, err := ms.db.ReadAll(responseNamespace)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "could not get all responses")
	}
	if len(gotResponses) == 0 {
		logrus.Info("no responses to get")
		return nil, nil
	}
	var stored []StoredResponse
	for responseKey, responseBytes := range gotResponses {
		var nextResponse StoredResponse
		if err = json.Unmarshal(responseBytes, &nextResponse); err == nil {
			stored = append(stored, nextResponse)
		} else {
			logrus.WithError(err).Errorf("could not unmarshal stored response: %s", responseKey)
		}
	}
	return stored, nil
}

func (ms *Storage) DeleteResponse(id string) error {
	if err := ms.db.Delete(responseNamespace, id); err != nil {
		return util.LoggingErrorMsgf(err, "could not delete response: %s", id)
	}
	return nil
}
