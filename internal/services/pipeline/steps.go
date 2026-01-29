package pipeline

import (
	"context"

	"github.com/aiservice/internal/models"
	"github.com/aiservice/internal/preprocessing"
	"github.com/aiservice/internal/providers"
	"github.com/firebase/genkit/go/ai"
)

const summarizePrompt = `
Тебе нужно следовать строго моей инструкции.
Ты получаешь набор "сырых" данных, которые нужно будет суметь обработать и ним выдать суммаризацию всего на доске.

1) Собери все элементы доски в общую композицию
2) Проанализируй то, что у тебя получилось
3) Дополнительно, посмотри изображение, которое я тебе дал - это скриншот доски, сверь себя с ним
4) Напиши обобщение того, к чему пришли пользователи на доске. К какому выводу/заключению.

Мне нужно, чтобы ты предоставил ответ в следующем формате:
1) Это должен быть текстовый элемент. Его модель следующая
type BaseElement struct {
	Id          string  json:"id"
	Type        string  json:"type" //text
	X           float32 json:"x"
	Y           float32 json:"y"
	Width       float32 json:"width"
	Height      float32 json:"height"
	Rotation    float32 json:"rotation"
	Fill        string  json:"fill,omitempty"
	Stroke      string  json:"stroke,omitempty"
	StrokeWidth int     json:"strokeWidth,omitempty"
	Content 	string  json:"content"
}
2) В поле Content напиши к чему пришли пользователи.
3) Сформируй правильное положение элемента относительно других, он должен находится в свободном месте.
4) Content - это html тип, который ограничен следующими тегами:
	Поддерживаемые теги: <p>, <br>, <strong>, <em>, <ul>, <ol>, <li>
`

const structurizePrompt = `
Тебе нужно следовать строго моей инструкции.
Ты получаешь набор "сырых" данных, которые нужно будет суметь обработать и ним выдать суммаризацию всего на доске.

1) Собери все элементы доски в общую композицию
2) Проанализируй то, что у тебя получилось
3) Дополнительно, посмотри изображение, которое я тебе дал - это скриншот доски, сверь себя с ним
4) Тебе нужно будет сделать файловую структуру проекта - это основной запрос пользователей.
5) Посмотри, к чему они пришли и начни формировать структуру.

Я жду от тебя ответ в таком формате:
type StructurizeResponse struct {
	RequestType    string json:"requestType"   // structurize
	AiTreeResponse string json:"aiTreeResponse" // дерево ASCII файлов
	File           File   json:"file"
}

type File struct {
	Name     string json:"name"
	Type     string json:"type" //doc, simple, graph(тогда поле children пустое), section(тогда содердит детей)
	Children []File json:"children"
}

6) В поле AiTreeResponse следует также добавить строковое представление получившейся файловой системы, в таком формате:
systemd─┬─AmneziaVPN-serv───AmneziaVPN-serv───{AmneziaVPN-serv}
        ├─ModemManager───File2
        ├─NetworkManager───File3
        ├─Dir1
		├   ├──File4
        ├─avahi-daemon───avahi-daemon
        ├─bluetoothd
`

// Preprocessor for transforming raw data into structured formats
var preprocessor = preprocessing.NewPreprocessor()

func newLlmSummarizeParts(req models.SummarizeRequest) ([]*ai.Part, error) {
	return preprocessor.PreprocessSummarizeRequest(req)
}

func newLlmStructurizeParts(req models.StructurizeRequest) ([]*ai.Part, error) {
	return preprocessor.PreprocessStructurizeRequest(req)
}

func newSummarizeStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		parts, err := newLlmSummarizeParts(state.AnalyzeRequest.SummarizeRequest)
		if err != nil {
			return err
		}
		resp, err := llm.Summarize(ctx, parts)
		if err != nil {
			return err
		}
		state.AnalyzeResponse.SummarizeResponse = fillSumRespWithMeta(resp, state)
		return nil
	}
}

func fillSumRespWithMeta(aiResp models.SummarizeResponse, state *PipelineState) models.SummarizeResponse {
	return models.SummarizeResponse{
		RequestID:   state.AnalyzeRequest.SummarizeRequest.RequestID,
		UserID:      state.AnalyzeRequest.SummarizeRequest.UserID,
		RequestType: models.SummarizeType,
		Element:     aiResp.Element,
	}
}

func newStructurizeStep(llm providers.LLMClient) Step {
	return func(ctx context.Context, state *PipelineState) error {
		parts, err := newLlmStructurizeParts(state.AnalyzeRequest.StructurizeRequest)
		if err != nil {
			return err
		}
		resp, err := llm.Structurize(ctx, parts)
		if err != nil {
			return err
		}
		state.AnalyzeResponse.StructurizeResponse = fillStructRespWithMeta(resp, state)
		return nil
	}
}

func fillStructRespWithMeta(aiResp models.StructurizeResponse, state *PipelineState) models.StructurizeResponse {
	return models.StructurizeResponse{
		RequestID:      state.AnalyzeRequest.SummarizeRequest.RequestID,
		UserID:         state.AnalyzeRequest.SummarizeRequest.UserID,
		RequestType:    models.StructurizeType,
		AiTreeResponse: aiResp.AiTreeResponse,
		File:           aiResp.File,
	}
}
