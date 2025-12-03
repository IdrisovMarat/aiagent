package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/IdrisovMarat/aiagent/internal/cloud"
	"github.com/IdrisovMarat/aiagent/internal/helpers"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("=== Консультация эндоваскулярного хирурга при остром ишемическом инсульте ===")
	fmt.Println()

	// Сбор данных пациента
	patient := helpers.CollectPatientData()

	fmt.Printf("\nСводка по пациенту:\n")
	fmt.Printf("ФИО пациента: %s\n", patient.Name)
	fmt.Printf("NIHSS: %d баллов, Время от начала: %s часов\n", patient.NIHSS, patient.OnsetTime)
	fmt.Printf("Возраст: %d лет, Пол: %s, mRS до инсульта: %d\n", patient.Age, patient.Gender, patient.MRS)
	fmt.Printf("Сознание: %s\n", patient.Consciousness)
	fmt.Printf("Неврологические симптомы: %v\n", patient.NeurologicSymptoms)
	fmt.Printf("Исследования: КТ нативная: %v, КТ ангиография: %v, КТ перфузия: %v\n",
		patient.CTRequired, patient.CTAngioRequired, patient.CTPerfusionRequired)

	// Если требуются исследования, сначала создаем их
	if patient.CTRequired || patient.CTAngioRequired || patient.CTPerfusionRequired {
		fmt.Println("\nПроведение назначенных исследований...")
		helpers.CreateAndUploadStudies(patient)
		fmt.Println("✓ Все исследования завершены и загружены в облако")
		fmt.Println("Можно запросить консультацию эндоваскулярного хирурга")
		fmt.Println()
	}

	// Формируем сообщение для AI агента
	message := helpers.FormatPatientMessage(patient)

	// Отправка запроса AI агенту
	fmt.Println("\nОтправка данных эндоваскулярному хирургу...")
	fmt.Print("Запрос консультации... ")
	executionID, err := cloud.StartWorkflowExecution(message)
	if err != nil {
		fmt.Printf("ОШИБКА\nОшибка: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("УСПЕХ\n\n")

	// Ожидание ответа с визуализацией
	fmt.Println("Эндоваскулярный хирург анализирует данные:")
	response, duration, err := cloud.WaitForAIResponse(executionID)
	if err != nil {
		fmt.Printf("\rОШИБКА: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\rКонсультация завершена! Время анализа: %.1fсек\n\n", duration.Seconds())

	// Вывод результата
	fmt.Println("Заключение эндоваскулярного хирурга:")
	fmt.Println("════════════════════════════════════════")
	fmt.Println(response)
	fmt.Println("════════════════════════════════════════")

	// Сохраняем в PDF
	pdfPath, err := helpers.SaveReportToPDFSimple(patient.Name, response)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("Файл успешно создан: %s\n", pdfPath)

	// Сохранение как Markdown
	err = os.WriteFile("report.md", []byte(response), 0644)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

}
