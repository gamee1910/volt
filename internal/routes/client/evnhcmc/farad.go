package evnhcmc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type FaradRequest struct {
	Token        string
	CustomerCode string
	FromDate     string
	ToDate       string
}

type FaradResponse struct {
	State string    `json:"state"`
	Alert string    `json:"alert"`
	Data  FaradData `json:"data"`
}

type FaradData struct {
	NumberOfDays int           `json:"soNgay"`
	Title        string        `json:"tieude"`
	DailyOutputs []DailyOutput `json:"sanluong_tungngay"`
}

type DailyOutput struct {
	Date                 string  `json:"ngay"`
	FullDate             string  `json:"ngayFull"`
	OffPeakIndex         float64 `json:"TD"`            // Chỉ số Thấp điểm
	StandardIndex        float64 `json:"BT"`            // Chỉ số Bình thường
	PeakIndex            float64 `json:"CD"`            // Chỉ số Cao điểm
	TotalIndex           float64 `json:"Tong"`          // Tổng chỉ số
	OffPeakOutput        string  `json:"sanluong_TD"`   // Sản lượng Thấp điểm
	StandardOutput       string  `json:"sanluong_BT"`   // Sản lượng Bình thường
	PeakOutput           string  `json:"sanluong_CD"`   // Sản lượng Cao điểm
	TotalOutput          string  `json:"sanluong_tong"` // Tổng sản lượng
	MultiplicationFactor float64 `json:"hsn"`           // Hệ số nhân (HSN / Scaling factor)
	ActiveStandardDeliv  string  `json:"p_giao_bt"`     // P giao Bình thường (Active Power Delivered)
	ActiveOffPeakDeliv   string  `json:"p_giao_td"`     // P giao Thấp điểm
	ActivePeakDeliv      string  `json:"p_giao_cd"`     // P giao Cao điểm
	TotalActiveDeliv     string  `json:"tong_p_giao"`   // Tổng P giao
	TotalReactiveDeliv   string  `json:"tong_q_giao"`   // Tổng Q giao (Reactive Power Delivered)
	MeasurementTimestamp string  `json:"thoidiemdo"`    // Thời điểm đo
	IsBilled             int     `json:"isChotHoaDon"`  // Trạng thái chốt hóa đơn (0/1)
}

func (c *EVNClient) GetDailyPowerUsageData(
	ctx context.Context,
	reqData FaradRequest,
) (*FaradResponse, error) {

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	if err := writer.WriteField("input_makh", reqData.CustomerCode); err != nil {
		return nil, err
	}

	if err := writer.WriteField("input_tungay", reqData.FromDate); err != nil {
		return nil, err
	}

	if err := writer.WriteField("input_denngay", reqData.ToDate); err != nil {
		return nil, err
	}

	if err := writer.WriteField("token", reqData.Token); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, dienNangURL, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"EVNHCMC request failed: status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	var result FaradResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf(
			"parse EVNHCMC response failed: %w",
			err,
		)
	}

	return &result, nil
}
