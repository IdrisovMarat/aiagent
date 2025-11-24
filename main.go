package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	gatewayURL   = "https://d5dprdl8m0emb0eqi0b1.svoluuab.apigw.yandexcloud.net/gogw"
	workflowsURL = "https://serverless-workflows.api.cloud.yandex.net/workflows/v1/execution"
	maxWaitTime  = 60 * time.Second
	pollInterval = 500 * time.Millisecond
	timeout      = 30 * time.Second
)

// Структуры для основного клиента
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

type PatientData struct {
	NIHSS               int      `json:"nihss"`
	OnsetTime           string   `json:"onset_time"`
	Age                 int      `json:"age"`
	Gender              string   `json:"gender"`
	MRS                 int      `json:"mrs"`
	NeurologicSymptoms  []string `json:"neurologic_symptoms"`
	Consciousness       string   `json:"consciousness"`
	CTRequired          bool     `json:"ct_required"`
	CTAngioRequired     bool     `json:"ct_angio_required"`
	CTPerfusionRequired bool     `json:"ct_perfusion_required"`
}

// Структуры для генератора исследований
type CTStudy struct {
	StudyID         string    `json:"study_id"`
	PatientID       string    `json:"patient_id"`
	StudyDate       time.Time `json:"study_date"`
	Conclusion      string    `json:"conclusion"`
	IschemicSign    string    `json:"ischemic_sign"`
	ASPECTS         int       `json:"aspects"`
	EarlySigns      []string  `json:"early_signs"`
	StrokeTerritory string    `json:"stroke_territory"`
}

type CTAngioStudy struct {
	StudyID           string    `json:"study_id"`
	PatientID         string    `json:"patient_id"`
	StudyDate         time.Time `json:"study_date"`
	Conclusion        string    `json:"conclusion"`
	OcclusionSite     string    `json:"occlusion_site"`
	Collaterals       string    `json:"collaterals"`
	MTTScore          int       `json:"mtt_score"`
	StenosisPercent   int       `json:"stenosis_percent"`
	AffectedTerritory string    `json:"affected_territory"`
	IsOcclusion       bool      `json:"is_occlusion"`
}

type CTPerfusionStudy struct {
	StudyID           string    `json:"study_id"`
	PatientID         string    `json:"patient_id"`
	StudyDate         time.Time `json:"study_date"`
	Conclusion        string    `json:"conclusion"`
	CoreInfarctML     float64   `json:"core_infarct_ml"`
	PenumbraML        float64   `json:"penumbra_ml"`
	MismatchRatio     float64   `json:"mismatch_ratio"`
	TMaxLesionVolume  float64   `json:"tmax_lesion_volume"`
	AffectedTerritory string    `json:"affected_territory"`
}

type StrokeTerritory struct {
	Name           string
	CTConclusions  []string
	OcclusionSites []string
	VesselType     string
	Symptoms       []string
}

var (
	territories = []StrokeTerritory{
		{
			Name: "СМА_слева",
			CTConclusions: []string{
				"Признаки острой ишемии в бассейне левой средней мозговой артерии",
				"Гиподенсивность паренхимы в бассейне левой СМА",
				"Острая ишемия левой гемисферы",
			},
			OcclusionSites: []string{"Проксимальный сегмент левой СМА (M1)", "Дистальный сегмент левой СМА (M2)"},
			VesselType:     "СМА",
			Symptoms:       []string{"Гемипарез правых конечностей", "Гемиплегия правых конечностей", "Афазия"},
		},
		{
			Name: "СМА_справа",
			CTConclusions: []string{
				"Признаки острой ишемии в бассейне правой средней мозговой артерии",
				"Гиподенсивность паренхимы в бассейне правой СМА",
				"Острая ишемия правой гемисферы",
			},
			OcclusionSites: []string{"Проксимальный сегмент правой СМА (M1)", "Дистальный сегмент правой СМА (M2)"},
			VesselType:     "СМА",
			Symptoms:       []string{"Гемипарез левых конечностей", "Гемиплегия левых конечностей", "Игнорирование левой стороны"},
		},
		{
			Name: "Вертебробазилярный",
			CTConclusions: []string{
				"Ишемические изменения в вертебробазилярном бассейне",
				"Острая ишемия ствола мозга и мозжечка",
			},
			OcclusionSites: []string{"Проксимальный сегмент базилярной артерии", "Дистальный сегмент базилярной артерии"},
			VesselType:     "ВББ",
			Symptoms:       []string{"Атаксия", "Диплопия", "Дисфагия", "Нарушение сознания"},
		},
	}

	collaterals = []string{
		"Коллатерали развиты хорошо",
		"Умеренное развитие коллатералей",
		"Коллатерали развиты слабо",
	}
)

type StudyGenerator struct {
	patientID   string
	territory   *StrokeTerritory
	patientData *PatientData // Используем ту же структуру PatientData
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("=== Консультация эндоваскулярного хирурга при остром ишемическом инсульте ===")
	fmt.Println()

	// Сбор данных пациента
	patient := collectPatientData()

	// Формируем сообщение для AI агента
	message := formatPatientMessage(patient)

	fmt.Printf("\nСводка по пациенту:\n")
	fmt.Printf("NIHSS: %d баллов, Время от начала: %s часов\n", patient.NIHSS, patient.OnsetTime)
	fmt.Printf("Возраст: %d лет, Пол: %s, mRS до инсульта: %d\n", patient.Age, patient.Gender, patient.MRS)
	fmt.Printf("Сознание: %s\n", patient.Consciousness)
	fmt.Printf("Неврологические симптомы: %v\n", patient.NeurologicSymptoms)
	fmt.Printf("Исследования: КТ нативная: %v, КТ ангиография: %v, КТ перфузия: %v\n",
		patient.CTRequired, patient.CTAngioRequired, patient.CTPerfusionRequired)

	fmt.Println("\nОтправка данных эндоваскулярному хирургу...")

	// Если требуются исследования, сначала создаем их
	if patient.CTRequired || patient.CTAngioRequired || patient.CTPerfusionRequired {
		fmt.Println("\nПроведение назначенных исследований...")
		createAndUploadStudies(patient)
		fmt.Println("✓ Все исследования завершены и загружены в облако")
		fmt.Println("Можно запросить консультацию эндоваскулярного хирурга")
		fmt.Println()
	}

	// Отправка запроса AI агенту
	fmt.Print("Запрос консультации... ")
	executionID, err := startWorkflowExecution(message)
	if err != nil {
		fmt.Printf("ОШИБКА\nОшибка: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("УСПЕХ\n\n")

	// Ожидание ответа с визуализацией
	fmt.Println("Эндоваскулярный хирург анализирует данные:")
	response, duration, err := waitForAIResponse(executionID)
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
}

// Функции основного клиента
func collectPatientData() *PatientData {
	var patient PatientData

	fmt.Println("Введите данные пациента:")

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

func createAndUploadStudies(patient *PatientData) {
	// Генерируем уникальный ID пациента
	patientID := fmt.Sprintf("patient-%d-%d", patient.Age, time.Now().Unix())

	fmt.Printf("Создание исследований для пациента %s...\n", patientID)
	fmt.Printf("Данные неврологического осмотра:\n")
	fmt.Printf("- NIHSS: %d баллов\n", patient.NIHSS)
	fmt.Printf("- Сознание: %s\n", patient.Consciousness)
	fmt.Printf("- Симптомы: %v\n", patient.NeurologicSymptoms)

	// Создаем генератор исследований
	generator := NewStudyGenerator(patientID, patient)

	fmt.Println("\nСоздание согласованных исследований...")
	fmt.Println(generator.GetClinicalCorrelation())

	// Генерируем запрошенные исследования
	if patient.CTRequired {
		fmt.Println("\n1. Создание КТ нативной...")
		ctStudy := generator.GenerateCTStudy()
		generator.UploadToStorage(ctStudy, "ct")
	}

	if patient.CTAngioRequired {
		fmt.Println("\n2. Создание КТ ангиографии...")
		ctAngioStudy := generator.GenerateCTAngioStudy()
		generator.UploadToStorage(ctAngioStudy, "ctangio")
	}

	if patient.CTPerfusionRequired {
		fmt.Println("\n3. Создание КТ перфузии...")
		ctPerfusionStudy := generator.GenerateCTPerfusionStudy()
		generator.UploadToStorage(ctPerfusionStudy, "ctperfusion")
	}
}

func formatPatientMessage(patient *PatientData) string {
	return fmt.Sprintf(`Пациент с острым ишемическим инсультом:
- NIHSS: %d баллов
- Время от начала симптомов: %s часов
- Возраст: %d лет
- Пол: %s
- mRS до инсульта: %d
- Сознание: %s
- Неврологические симптомы: %v
- Проведенные исследования: КТ нативная: %v, КТ ангиография: %v, КТ перфузия: %v

Проанализируйте данные пациента и дайте рекомендации по лечению:
1. Возможность проведения системного тромболизиса
2. Показания к эндоваскулярному thrombectomy
3. Прогноз и ожидаемые результаты
4. Дополнительные рекомендации`,
		patient.NIHSS, patient.OnsetTime, patient.Age, patient.Gender, patient.MRS,
		patient.Consciousness, patient.NeurologicSymptoms,
		patient.CTRequired, patient.CTAngioRequired, patient.CTPerfusionRequired)
}

// Функции генератора исследований
func NewStudyGenerator(patientID string, patientData *PatientData) *StudyGenerator {
	// Выбираем территорию на основе симптомов
	var territory *StrokeTerritory

	// Определяем территорию по симптомам
	for i := range territories {
		for _, symptom := range patientData.NeurologicSymptoms {
			for _, terrSymptom := range territories[i].Symptoms {
				if symptom == terrSymptom {
					territory = &territories[i]
					break
				}
			}
			if territory != nil {
				break
			}
		}
		if territory != nil {
			break
		}
	}

	// Если не нашли по симптомам, выбираем случайную
	if territory == nil {
		territory = &territories[rand.Intn(len(territories))]
	}

	return &StudyGenerator{
		patientID:   patientID,
		territory:   territory,
		patientData: patientData,
	}
}

func (sg *StudyGenerator) GenerateCTStudy() *CTStudy {
	// ASPECTS зависит от тяжести инсульта
	var aspects int
	if sg.patientData.NIHSS > 15 || sg.patientData.Consciousness == "Кома" {
		aspects = rand.Intn(3) + 3 // 3-5 при тяжелом инсульте
	} else if sg.patientData.NIHSS > 10 {
		aspects = rand.Intn(3) + 5 // 5-7 при среднетяжелом
	} else {
		aspects = rand.Intn(3) + 7 // 7-10 при легком
	}

	// Ранние признаки
	earlySigns := []string{}
	if aspects <= 7 { // Только при значимом поражении
		signs := []string{"Гиподенсивность паренхимы", "Стирание границ серого и белого вещества"}
		for i := 0; i < rand.Intn(2)+1; i++ {
			earlySigns = append(earlySigns, signs[rand.Intn(len(signs))])
		}
	}

	return &CTStudy{
		StudyID:         fmt.Sprintf("CT-%d", rand.Intn(10000)),
		PatientID:       sg.patientID,
		StudyDate:       time.Now(),
		Conclusion:      sg.territory.CTConclusions[rand.Intn(len(sg.territory.CTConclusions))],
		IschemicSign:    "Гиподенсивность паренхимы",
		ASPECTS:         aspects,
		EarlySigns:      earlySigns,
		StrokeTerritory: sg.territory.Name,
	}
}

func (sg *StudyGenerator) GenerateCTAngioStudy() *CTAngioStudy {
	// Определяем тип поражения артерии на основе тяжести инсульта
	var isOcclusion bool
	var stenosisPercent int

	if sg.patientData.NIHSS > 15 || sg.patientData.Consciousness == "Кома" {
		// Тяжелый инсульт - почти всегда окклюзия
		isOcclusion = true
		stenosisPercent = 100
	} else if sg.patientData.NIHSS > 10 {
		// Среднетяжелый - 80% окклюзия, 20% стеноз 90%
		if rand.Float32() < 0.8 {
			isOcclusion = true
			stenosisPercent = 100
		} else {
			isOcclusion = false
			stenosisPercent = 90
		}
	} else {
		// Легкий - 50% окклюзия, 50% стеноз 90%
		if rand.Float32() < 0.5 {
			isOcclusion = true
			stenosisPercent = 100
		} else {
			isOcclusion = false
			stenosisPercent = 90
		}
	}

	conclusion := ""
	if isOcclusion {
		conclusion = fmt.Sprintf("Полная окклюзия: %s",
			sg.territory.OcclusionSites[rand.Intn(len(sg.territory.OcclusionSites))])
	} else {
		conclusion = fmt.Sprintf("Критический стеноз %d%%: %s", stenosisPercent,
			sg.territory.OcclusionSites[rand.Intn(len(sg.territory.OcclusionSites))])
	}

	return &CTAngioStudy{
		StudyID:           fmt.Sprintf("CTA-%d", rand.Intn(10000)),
		PatientID:         sg.patientID,
		StudyDate:         time.Now(),
		Conclusion:        conclusion,
		OcclusionSite:     sg.territory.OcclusionSites[rand.Intn(len(sg.territory.OcclusionSites))],
		Collaterals:       collaterals[rand.Intn(len(collaterals))],
		MTTScore:          rand.Intn(3) + 1,
		StenosisPercent:   stenosisPercent,
		AffectedTerritory: sg.territory.Name,
		IsOcclusion:       isOcclusion,
	}
}

func (sg *StudyGenerator) GenerateCTPerfusionStudy() *CTPerfusionStudy {
	// Объем поражения зависит от NIHSS
	var coreInfarct, penumbra float64

	if sg.patientData.NIHSS > 15 {
		coreInfarct = 30 + rand.Float64()*40 // 30-70 мл
		penumbra = 40 + rand.Float64()*60    // 40-100 мл
	} else if sg.patientData.NIHSS > 10 {
		coreInfarct = 15 + rand.Float64()*25 // 15-40 мл
		penumbra = 25 + rand.Float64()*35    // 25-60 мл
	} else {
		coreInfarct = 5 + rand.Float64()*15 // 5-20 мл
		penumbra = 10 + rand.Float64()*20   // 10-30 мл
	}

	mismatchRatio := penumbra / coreInfarct
	if coreInfarct < 1 {
		mismatchRatio = 3.0
	}

	conclusion := ""
	if mismatchRatio > 1.8 && coreInfarct < 70 {
		conclusion = "Благоприятный профиль для реперфузии"
	} else if coreInfarct > 70 {
		conclusion = "Большой объем ядра инфаркта, неблагоприятный прогноз"
	} else {
		conclusion = "Ограниченные возможности реперфузии"
	}

	return &CTPerfusionStudy{
		StudyID:           fmt.Sprintf("CTP-%d", rand.Intn(10000)),
		PatientID:         sg.patientID,
		StudyDate:         time.Now(),
		Conclusion:        conclusion,
		CoreInfarctML:     coreInfarct,
		PenumbraML:        penumbra,
		MismatchRatio:     mismatchRatio,
		TMaxLesionVolume:  penumbra + coreInfarct,
		AffectedTerritory: sg.territory.Name,
	}
}

func (sg *StudyGenerator) GetClinicalCorrelation() string {
	correlation := "Клинико-радиологическая корреляция:\n"
	correlation += fmt.Sprintf("- NIHSS: %d баллов\n", sg.patientData.NIHSS)
	correlation += fmt.Sprintf("- Сознание: %s\n", sg.patientData.Consciousness)
	correlation += fmt.Sprintf("- Территория инсульта: %s\n", sg.territory.Name)
	correlation += fmt.Sprintf("- Типичные симптомы: %v\n", sg.territory.Symptoms)
	correlation += fmt.Sprintf("- Выявленные симптомы: %v\n", sg.patientData.NeurologicSymptoms)
	return correlation
}

func (sg *StudyGenerator) UploadToStorage(study interface{}, studyType string) error {
	jsonData, err := json.MarshalIndent(study, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("Загружаем в папку %s:\n", studyType)
	fmt.Println(string(jsonData))
	fmt.Println("---")

	time.Sleep(1 * time.Second)
	return nil
}

// Остальные функции основного клиента
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
		"[▰▱▱▱▱▱▱▱▱] Анализ NIHSS...",
		"[▰▰▱▱▱▱▱▱▱] Оценка временного окна...",
		"[▰▰▰▱▱▱▱▱▱] Изучение КТ данных...",
		"[▰▰▰▰▱▱▱▱▱] Анализ ангиографии...",
		"[▰▰▰▰▰▱▱▱▱] Оценка перфузии...",
		"[▰▰▰▰▰▰▱▱▱] Определение показаний...",
		"[▰▰▰▰▰▰▰▱▱] Формирование рекомендаций...",
		"[▰▰▰▰▰▰▰▰▱] Подготовка заключения...",
		"[▰▰▰▰▰▰▰▰▰] Завершение анализа...",
	}
	animationIndex := 0
	startTime := time.Now()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Обновляем анимацию
			fmt.Printf("\r%s", animations[animationIndex])
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
