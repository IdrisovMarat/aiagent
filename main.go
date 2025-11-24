package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	gatewayURL   = "https://d5dprdl8m0emb0eqi0b1.svoluuab.apigw.yandexcloud.net/gogw"
	workflowsURL = "https://serverless-workflows.api.cloud.yandex.net/workflows/v1/execution"
	maxWaitTime  = 30 * time.Second
	pollInterval = 500 * time.Millisecond
	timeout      = 30 * time.Second
)

type GatewayRequest struct {
	Message string `json:"message"`
}

type GatewayResponse struct {
	ExecutionID string `json:"executionId"`
}

type WorkflowExecution struct {
	Execution struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Result struct {
			ResultJSON string `json:"resultJson"`
		} `json:"result"`
		Error string `json:"error,omitempty"`
	} `json:"execution"`
}

type AIAgentResponse struct {
	AiStudioAgentCall struct {
		Result string `json:"Result"`
	} `json:"ai-studio-agent-call"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ai-client <message>")
		fmt.Println("Example: ai-client 'Привет! Как тебя зовут?'")
		os.Exit(1)
	}

	message := os.Args[1]

	fmt.Printf("Вопрос: %s\n\n", message)

	// Отправка запроса
	fmt.Print("Отправка запроса... ")
	executionID, err := startWorkflowExecution(message)
	if err != nil {
		fmt.Printf("ОШИБКА\nОшибка: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("УСПЕХ\n\n")

	// Ожидание ответа с визуализацией
	fmt.Println("AI агент генерирует ответ:")
	response, duration, err := waitForAIResponse(executionID)
	if err != nil {
		fmt.Printf("\rОШИБКА: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\rГотово! Время ответа: %.1fсек\n\n", duration.Seconds())

	// Вывод результата
	fmt.Println("Ответ AI агента:")
	fmt.Println("────────────────────")
	fmt.Println(response)
	fmt.Println("────────────────────")
}

func startWorkflowExecution(message string) (string, error) {
	request := GatewayRequest{Message: message}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	resp, err := http.Post(gatewayURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("gateway request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gateway error %d: %s", resp.StatusCode, string(body))
	}

	var gatewayResp GatewayResponse
	if err := json.NewDecoder(resp.Body).Decode(&gatewayResp); err != nil {
		return "", fmt.Errorf("decode gateway response: %w", err)
	}

	if gatewayResp.ExecutionID == "" {
		return "", fmt.Errorf("empty execution ID")
	}

	return gatewayResp.ExecutionID, nil
}

func waitForAIResponse(executionID string) (string, time.Duration, error) {
	token, err := getIAMToken()
	if err != nil {
		return "", 0, fmt.Errorf("get IAM token: %w", err)
	}

	animations := []string{
		"[▰▱▱▱▱▱▱▱▱]",
		"[▰▰▱▱▱▱▱▱▱]",
		"[▰▰▰▱▱▱▱▱▱]",
		"[▰▰▰▰▱▱▱▱▱]",
		"[▰▰▰▰▰▱▱▱▱]",
		"[▰▰▰▰▰▰▱▱▱]",
		"[▰▰▰▰▰▰▰▱▱]",
		"[▰▰▰▰▰▰▰▰▱]",
		"[▰▰▰▰▰▰▰▰▰]",
		"[▰▰▰▰▰▰▰▰▰]",
	}
	animationIndex := 0
	startTime := time.Now()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Обновляем анимацию
			fmt.Printf("\r%s Обработка запроса...", animations[animationIndex])
			animationIndex = (animationIndex + 1) % len(animations)

			// Проверяем статус
			execution, err := getExecutionStatus(executionID, token)
			if err != nil {
				continue
			}

			switch execution.Execution.Status {
			case "FINISHED":
				duration := time.Since(startTime)
				result, err := parseAIResult(execution.Execution.Result.ResultJSON)
				return result, duration, err
			case "FAILED":
				return "", 0, fmt.Errorf("workflow failed: %s", execution.Execution.Error)
			case "RUNNING":
				continue
			default:
				continue
			}

		case <-time.After(maxWaitTime):
			return "", 0, fmt.Errorf("timeout: AI агент не ответил в течение %v", maxWaitTime)
		}
	}
}

func getExecutionStatus(executionID, token string) (*WorkflowExecution, error) {
	url := fmt.Sprintf("%s/%s", workflowsURL, executionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status check failed: %d", resp.StatusCode)
	}

	var execution WorkflowExecution
	if err := json.NewDecoder(resp.Body).Decode(&execution); err != nil {
		return nil, err
	}

	return &execution, nil
}

func parseAIResult(resultJSON string) (string, error) {
	if resultJSON == "" {
		return "", fmt.Errorf("empty result JSON")
	}

	var aiResp AIAgentResponse
	if err := json.Unmarshal([]byte(resultJSON), &aiResp); err != nil {
		return "", fmt.Errorf("parse AI response: %w", err)
	}

	if aiResp.AiStudioAgentCall.Result == "" {
		return "", fmt.Errorf("empty AI response")
	}

	return aiResp.AiStudioAgentCall.Result, nil
}

func getIAMToken() (string, error) {
	cmd := exec.Command("yc", "iam", "create-token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("execute yc command: %w", err)
	}
	return string(bytes.TrimSpace(output)), nil
}
