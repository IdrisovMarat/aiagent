package helpers

import (
	"fmt"
	"time"
)

func CollectPatientData() *PatientData {
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

func FormatPatientMessage(patient *PatientData) string {
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

func CreateAndUploadStudies(patient *PatientData) {
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
