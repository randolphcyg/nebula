package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"nebula/internal/config"
)

type Service struct {
	client *http.Client
}

func NewService() *Service {
	cfg := config.Get()
	return &Service{
		client: &http.Client{Timeout: cfg.Wireshark.GetTimeout()},
	}
}

func (s *Service) GetWiresharkVersion() (string, error) {
	cfg := config.Get()
	respData, err := s.doGet(cfg.Wireshark.BaseURL + "/version/wireshark")
	if err != nil {
		return "", err
	}

	var data struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(respData, &data); err != nil {
		return "", fmt.Errorf("解析版本信息失败：%w", err)
	}

	return data.Version, nil
}

func (s *Service) GetPacketsByPage(fileName string, page int, size int, bpfFilter string) (string, error) {
	cfg := config.Get()
	reqBody := getByPageRequest{
		baseRequest: baseRequest{
			Filepath:  toContainerPath(fileName),
			IsDebug:   false,
			IgnoreErr: true,
			BpfFilter: bpfFilter,
		},
		Page: page,
		Size: size,
	}

	respData, err := s.doPost(cfg.Wireshark.BaseURL+"/frames/page", reqBody)
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

func (s *Service) GetAllFrames(filepath string, bpfFilter string) (string, error) {
	cfg := config.Get()
	reqBody := baseRequest{
		Filepath:  toContainerPath(filepath),
		IsDebug:   false,
		IgnoreErr: true,
		BpfFilter: bpfFilter,
	}

	respData, err := s.doPost(cfg.Wireshark.BaseURL+"/frames/all", reqBody)
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

func (s *Service) GetPacketDetail(filepath string, index int) (string, error) {
	cfg := config.Get()
	reqBody := getByIdxsRequest{
		baseRequest: baseRequest{
			Filepath:  toContainerPath(filepath),
			IsDebug:   false,
			IgnoreErr: true,
		},
		FrameIdxs: []int{index},
	}

	respData, err := s.doPost(cfg.Wireshark.BaseURL+"/frames/idxs", reqBody)
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

func (s *Service) GetPacketHex(filepath string, index int) (string, error) {
	cfg := config.Get()
	reqBody := getHexReq{
		baseRequest: baseRequest{
			Filepath:  toContainerPath(filepath),
			IsDebug:   false,
			IgnoreErr: true,
		},
		FrameIdx: index,
	}

	respData, err := s.doPost(cfg.Wireshark.BaseURL+"/frames/hex", reqBody)
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

func (s *Service) FollowStream(filepath string, bpfFilter string, protocol string) (string, error) {
	cfg := config.Get()
	reqBody := getStreamReq{
		baseRequest: baseRequest{
			Filepath:  toContainerPath(filepath),
			IsDebug:   false,
			IgnoreErr: true,
			BpfFilter: bpfFilter,
		},
		Protocol: protocol,
	}

	respData, err := s.doPost(cfg.Wireshark.BaseURL+"/frames/stream", reqBody)
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

func (s *Service) GetInterfaces() (string, error) {
	cfg := config.Get()
	respData, err := s.doGet(cfg.Wireshark.BaseURL + "/interfaces")
	if err != nil {
		return "", err
	}
	return string(respData), nil
}

func toContainerPath(fullPath string) string {
	cfg := config.Get()
	slashPath := filepath.ToSlash(fullPath)

	idx := strings.LastIndex(slashPath, "/pcaps/")
	if idx != -1 {
		relativePath := slashPath[idx+len("/pcaps/"):]
		return cfg.Wireshark.ContainerMountPath + relativePath
	}

	dir := filepath.Base(filepath.Dir(slashPath))
	base := filepath.Base(slashPath)
	if dir == "pcaps" || dir == "." || dir == "/" {
		return cfg.Wireshark.ContainerMountPath + base
	}

	return cfg.Wireshark.ContainerMountPath + dir + "/" + base
}

func (s *Service) doGet(url string) ([]byte, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 wireshark 失败：%v", err)
	}
	defer resp.Body.Close()
	return handleResponse(resp)
}

func (s *Service) doPost(url string, body interface{}) ([]byte, error) {
	jsonBytes, _ := json.Marshal(body)
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("请求 wireshark 失败：%v", err)
	}
	defer resp.Body.Close()
	return handleResponse(resp)
}

func handleResponse(resp *http.Response) ([]byte, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	var apiRes apiResponse
	if err := json.Unmarshal(bodyBytes, &apiRes); err != nil {
		return nil, fmt.Errorf("解析 wireshark 响应失败：%v", err)
	}

	if apiRes.Code != 0 {
		return nil, fmt.Errorf("wireshark 错误 [%d]: %s (err: %s)", apiRes.Code, apiRes.Msg, apiRes.Error)
	}

	return apiRes.Data, nil
}
