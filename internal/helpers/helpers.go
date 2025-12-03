package helpers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

func CollectPatientData() *PatientData {
	var patient PatientData

	fmt.Println("Введите данные пациента:")

	// Генерируем уникальный ID пациента
	patient.ID = fmt.Sprintf("patient-%d-%d", patient.Age, time.Now().Unix())

	fmt.Print("\nВведите фамилию пациента: ")
	// Используем bufio для чтения строки целиком
	reader := bufio.NewReader(os.Stdin)
	nameInput, _ := reader.ReadString('\n')
	nameInput = strings.TrimSpace(nameInput)

	if nameInput == "" {
		fmt.Println("\nИспользуются тестовые данные пациента...")

		patient.Name = "Петров"
		patient.Age = 66
		patient.Gender = "м"
		patient.OnsetTime = "3"
		patient.Consciousness = "Оглушение"
		patient.NIHSS = 10
		patient.MRS = 0
		patient.NeurologicSymptoms = []string{"Гемиплегия левых конечностей"}
		patient.CTRequired = true
		patient.CTAngioRequired = true
		patient.CTPerfusionRequired = false
		patient.CT = "не проведено"
		patient.CTAngio = "не проведено"
		patient.CTPerfusion = "не проведено"

		return &patient
	}

	// Если имя введено, продолжаем сбор данных обычным образом
	patient.Name = nameInput

	fmt.Print("NIHSS (0-42): ")
	fmt.Scan(&patient.NIHSS)

	fmt.Print("Время от начала симптомов (часы): ")
	fmt.Scan(&patient.OnsetTime)

	fmt.Print("Возраст: ")
	fmt.Scan(&patient.Age)

	fmt.Print("Пол (м/ж): ")
	fmt.Scan(&patient.Gender)

	fmt.Print("mRS до инсульта (0-5): ")
	fmt.Scan(&patient.MRS)

	// Оценка сознания
	fmt.Println("\nУровень сознания:")
	fmt.Println("1 - Ясное")
	fmt.Println("2 - Оглушение")
	fmt.Println("3 - Сопор")
	fmt.Println("4 - Кома (ШКГ 3-5 баллов)")
	fmt.Print("Выберите вариант (1-4): ")
	var consciousness int
	fmt.Scan(&consciousness)

	switch consciousness {
	case 1:
		patient.Consciousness = "Ясное"
	case 2:
		patient.Consciousness = "Оглушение"
	case 3:
		patient.Consciousness = "Сопор"
	case 4:
		patient.Consciousness = "Кома"
	default:
		patient.Consciousness = "Ясное"
	}

	// Неврологические симптомы
	fmt.Println("\nНеврологические симптомы (выберите номера, через запятую):")
	fmt.Println("1 - Гемипарез правых конечностей")
	fmt.Println("2 - Гемипарез левых конечностей")
	fmt.Println("3 - Гемиплегия правых конечностей")
	fmt.Println("4 - Гемиплегия левых конечностей")
	fmt.Println("5 - Афазия")
	fmt.Println("6 - Дизартрия")
	fmt.Println("7 - Лицевой парез")
	fmt.Println("8 - Гемигипестезия")
	fmt.Println("9 - Гемианопсия")
	fmt.Println("10 - Атаксия")
	fmt.Println("11 - Диплопия")
	fmt.Println("12 - Дисфагия")
	fmt.Print("Введите номера симптомов: ")

	var symptomsInput string
	fmt.Scan(&symptomsInput)

	symptomsMap := map[string]string{
		"1":  "Гемипарез правых конечностей",
		"2":  "Гемипарез левых конечностей",
		"3":  "Гемиплегия правых конечностей",
		"4":  "Гемиплегия левых конечностей",
		"5":  "Афазия",
		"6":  "Дизартрия",
		"7":  "Лицевой парез",
		"8":  "Гемигипестезия",
		"9":  "Гемианопсия",
		"10": "Атаксия",
		"11": "Диплопия",
		"12": "Дисфагия",
	}

	patient.NeurologicSymptoms = []string{}
	for i := 0; i < len(symptomsInput); i++ {
		if string(symptomsInput[i]) != "," {
			if symptom, exists := symptomsMap[string(symptomsInput[i])]; exists {
				patient.NeurologicSymptoms = append(patient.NeurologicSymptoms, symptom)
			}
		}
	}

	fmt.Print("\nНазначить исследования:\n")
	fmt.Print("КТ нативная (да/нет): ")
	var ct string
	fmt.Scan(&ct)
	patient.CTRequired = (ct == "да")

	fmt.Print("КТ ангиография (да/нет): ")
	var ctAngio string
	fmt.Scan(&ctAngio)
	patient.CTAngioRequired = (ctAngio == "да")

	fmt.Print("КТ перфузия (да/нет): ")
	var ctPerf string
	fmt.Scan(&ctPerf)
	patient.CTPerfusionRequired = (ctPerf == "да")

	return &patient
}

func FormatPatientMessage(patient *PatientData) string {
	return fmt.Sprintf(`Пациент с острым ишемическим инсультом:
- ID: %s
- ФИО: %s
- NIHSS: %d баллов
- Время от начала симптомов: %s часов
- Возраст: %d лет
- Пол: %s
- mRS до инсульта: %d
- Сознание: %s
- Неврологические симптомы: %v
- КТ нативная: %s
- КТ ангиография: %s
- КТ перфузия: %s
`,
		patient.ID, patient.Name, patient.NIHSS, patient.OnsetTime, patient.Age, patient.Gender, patient.MRS,
		patient.Consciousness, patient.NeurologicSymptoms,
		patient.CT, patient.CTAngio, patient.CTPerfusion)

	// Проанализируйте данные пациента и дайте рекомендации по лечению:
	// 1. Возможность проведения системного тромболизиса
	// 2. Показания к тромбектомии
	// 3. Прогноз и ожидаемые результаты
	// 4. Дополнительные рекомендации
}

func CreateAndUploadStudies(patient *PatientData) {
	fmt.Printf("Создание исследований для пациента %s...\n", patient.ID)
	// Создаем генератор исследований
	generator := NewStudyGenerator(patient.ID, patient)

	// Генерируем запрошенные исследования
	if patient.CTRequired {
		fmt.Println("\n1. Создание КТ нативной...")
		ctStudy := generator.GenerateCTStudy()
		patient.CT = ctStudy.StudyID
		generator.UploadToStorage(ctStudy, "ct")
	}

	if patient.CTAngioRequired {
		fmt.Println("\n2. Создание КТ ангиографии...")
		ctAngioStudy := generator.GenerateCTAngioStudy()
		patient.CTAngio = ctAngioStudy.StudyID
		generator.UploadToStorage(ctAngioStudy, "ctangio")
	}

	if patient.CTPerfusionRequired {
		fmt.Println("\n3. Создание КТ перфузии...")
		ctPerfusionStudy := generator.GenerateCTPerfusionStudy()
		patient.CTPerfusion = ctPerfusionStudy.StudyID
		generator.UploadToStorage(ctPerfusionStudy, "ctperfusion")
	}
}

// Вспомогательная функция для очистки имени файла
func sanitizeFileName(name string) string {
	// Удаляем недопустимые символы
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Ограничиваем длину
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

// Альтернативная функция с более простым подходом (без внешних библиотек для Markdown)
func SaveReportToPDFSimple(patientName, markdownContent string) (string, error) {
	cleanPatientName := sanitizeFileName(patientName)
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.pdf", cleanPatientName, timestamp)
	pdfPath := filepath.Join(".", filename)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Заголовок
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("Медицинский отчет: %s", patientName))
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 10, fmt.Sprintf("Дата создания: %s", time.Now().Format("02.01.2006 15:04:05")))
	pdf.Ln(20)

	// Преобразуем Markdown в простой текст
	pdf.SetFont("Arial", "", 11)
	lines := strings.Split(markdownContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			pdf.Ln(5)
			continue
		}

		// Базовая обработка Markdown
		if strings.HasPrefix(line, "# ") {
			pdf.SetFont("Arial", "B", 14)
			pdf.Cell(0, 10, strings.TrimPrefix(line, "# "))
			pdf.Ln(10)
			pdf.SetFont("Arial", "", 11)
		} else if strings.HasPrefix(line, "## ") {
			pdf.SetFont("Arial", "B", 12)
			pdf.Cell(0, 10, strings.TrimPrefix(line, "## "))
			pdf.Ln(8)
			pdf.SetFont("Arial", "", 11)
		} else if strings.HasPrefix(line, "- ") {
			pdf.Cell(5, 6, "")
			pdf.Cell(0, 6, "• "+strings.TrimPrefix(line, "- "))
			pdf.Ln(6)
		} else {
			// Обычный текст с переносом
			pdf.MultiCell(0, 6, line, "", "", false)
			pdf.Ln(2)
		}
	}

	err := pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения PDF: %w", err)
	}

	fmt.Printf("PDF отчет сохранен: %s\n", pdfPath)
	return pdfPath, nil
}
