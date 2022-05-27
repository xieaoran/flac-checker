package remote

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/xieaoran/flac-checker/models"
)

type MORATrackResultList struct {
	Results []*MORATrackResult `json:"list"`
	Total   int                `json:"total"`
}

type MORATrackResult struct {
	DistNumber     string `json:"distPartNo"`
	MaterialNumber int    `json:"materialNo"`
	ArtistName     string `json:"artistName"`
	TrackTitle     string `json:"trackTitle"`
	PackageID      string `json:"packageId"`
	PackageNumber  int    `json:"packageNo"`
	PackageTitle   string `json:"packageTitle"`
	PackagePage    string `json:"packagePage"`
	SampleRate     int    `json:"samplingFreq"`
	BitDepth       string `json:"bitPerSample"`
	Price          int    `json:"price"`
	StartDate      string `json:"startDate"`
}

func (moraResult *MORATrackResult) Marshal() (*models.FileMeta, error) {
	bitDepth, parseError := strconv.ParseInt(moraResult.BitDepth, 10, 64)
	if parseError != nil {
		return nil, parseError
	}

	return &models.FileMeta{
		FilePath:   moraResult.PackagePage,
		Title:      moraResult.TrackTitle,
		Artist:     moraResult.ArtistName,
		Album:      moraResult.PackageTitle,
		SampleRate: moraResult.SampleRate,
		BitDepth:   int(bitDepth),
	}, nil
}

type MORAResponseHead struct {
	Message     string `json:"message"`
	SuccessFlag string `json:"successFlg"`
}

type MORAResponseData struct {
	TrackResults *MORATrackResultList `json:"trackResult"`
}

type MORAResponse struct {
	Head *MORAResponseHead `json:"head"`
	Data *MORAResponseData `json:"data"`
}

const MORAURL string = "https://mora.jp/search/getResult"

type MORAProvider struct {
	httpClient *http.Client
}

func NewMORAProvider() *MORAProvider {
	return &MORAProvider{httpClient: new(http.Client)}
}

func (*MORAProvider) Name() string {
	return "mora"
}

func (provider *MORAProvider) Search(localMeta *models.FileMeta) ([]*models.FileMeta, error) {
	request, requestError := http.NewRequest(http.MethodGet, MORAURL, nil)
	if requestError != nil {
		return nil, requestError
	}

	query := request.URL.Query()
	query.Add("keyWord", localMeta.Title+" "+localMeta.Artist)
	query.Add("onlyHires", "1")
	request.URL.RawQuery = query.Encode()

	responseRaw, httpError := provider.httpClient.Do(request)
	if httpError != nil {
		return nil, fmt.Errorf("HTTP failed, request.URL[%s], httpError.Error[%s]",
			request.URL.String(), httpError.Error())
	}
	defer responseRaw.Body.Close()

	responseJSON, ioError := ioutil.ReadAll(responseRaw.Body)
	if ioError != nil {
		return nil, ioError
	}

	response := new(MORAResponse)
	jsonError := json.Unmarshal(responseJSON, response)
	if jsonError != nil {
		return nil, fmt.Errorf("json.Unmarshal failed, "+
			"request.URL[%s], responseJSON[%s], jsonError.Error[%s]",
			request.URL.String(), string(responseJSON), jsonError.Error())
	}
	if response.Data == nil || response.Data.TrackResults == nil {
		return nil, fmt.Errorf("response nil, "+
			"request.URL[%s], responseJSON[%s]", request.URL.String(), string(responseJSON))
	}

	var results []*models.FileMeta
	for _, trackResult := range response.Data.TrackResults.Results {
		fileMeta, marshalError := trackResult.Marshal()
		if marshalError != nil {
			return nil, fmt.Errorf("trackResult.Marshal failed, "+
				"trackResult[%+v], marshalError.Error[%s]", trackResult, marshalError.Error())
		}

		if fileMeta.BitDepth > localMeta.BitDepth || fileMeta.SampleRate > localMeta.SampleRate {
			results = append(results, fileMeta)
		}
	}

	return results, nil
}
