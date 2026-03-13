package analyzer

import "encoding/json"

// ======================= 请求数据结构 =======================

type baseRequest struct {
	Filepath  string `json:"filepath"`
	IsDebug   bool   `json:"isDebug"`
	IgnoreErr bool   `json:"ignoreErr"`
	BpfFilter string `json:"bpfFilter"`
}

type getByPageRequest struct {
	baseRequest
	Page int `json:"page"`
	Size int `json:"size"`
}

type getByIdxsRequest struct {
	baseRequest
	FrameIdxs []int `json:"frameIdxs"`
}

type getHexReq struct {
	baseRequest
	FrameIdx int `json:"frameIdx"`
}

type getStreamReq struct {
	baseRequest
	Protocol string `json:"protocol"`
}

// ======================= 响应数据结构 =======================

type apiResponse struct {
	Code  int             `json:"code"`
	Msg   string          `json:"msg"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}
