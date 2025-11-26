package helpers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

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
