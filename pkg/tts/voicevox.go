package tts

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hybridgroup/voicevox/pkg/voicevox"
)

// Voicevox implements tts.Speaker using VOICEVOX core for Japanese speech.
type Voicevox struct {
	style voicevox.StyleID
	syn   voicevox.Synthesizer
	gpu   bool

	speedScale      *float64
	pitchScale      *float64
	intonationScale *float64
	volumeScale     *float64
}

// NewVoicevox returns a Voicevox speaker that uses the given style id.
func NewVoicevox(style uint32) *Voicevox {
	return &Voicevox{style: voicevox.StyleID(style)}
}

// UseGPU makes Connect load the CUDA libraries and require the GPU. Otherwise VOICEVOX picks the mode.
func (v *Voicevox) UseGPU(gpu bool) { v.gpu = gpu }

// SetSpeedScale sets the speaking speed. The default is 1.0.
func (v *Voicevox) SetSpeedScale(s float64) { v.speedScale = &s }

// SetPitchScale sets the pitch offset. The default is 0.0.
func (v *Voicevox) SetPitchScale(s float64) { v.pitchScale = &s }

// SetIntonationScale sets the intonation strength. The default is 1.0.
func (v *Voicevox) SetIntonationScale(s float64) { v.intonationScale = &s }

// SetVolumeScale sets the volume. The default is 1.0.
func (v *Voicevox) SetVolumeScale(s float64) { v.volumeScale = &s }

// Connect loads VOICEVOX from datadir, which uses the layout created by the VOICEVOX downloader.
func (v *Voicevox) Connect(datadir string) error {
	if err := voicevox.Load(filepath.Join(datadir, "c_api", "lib")); err != nil {
		return fmt.Errorf("voicevox: unable to load voicevox_core: %w", err)
	}

	opts := voicevox.DefaultInitializeOptions()
	if v.gpu {
		libs := filepath.Join(datadir, "additional_libraries")
		if err := voicevox.LoadAdditionalLibraries(libs); err != nil {
			return fmt.Errorf("voicevox: unable to load CUDA libraries from %s: %w", libs, err)
		}
		opts.AccelerationMode = voicevox.AccelerationModeGPU
	}

	ort, err := voicevox.LoadOnnxruntime(filepath.Join(datadir, "onnxruntime", "lib"))
	if err != nil {
		return fmt.Errorf("voicevox: unable to load ONNX Runtime: %w", err)
	}

	ojt, err := voicevox.NewOpenJtalk(filepath.Join(datadir, "dict", "open_jtalk_dic_utf_8-1.11"))
	if err != nil {
		return fmt.Errorf("voicevox: unable to load Open JTalk dictionary: %w", err)
	}

	v.syn, err = voicevox.NewSynthesizer(ort, ojt, opts)
	ojt.Delete()
	switch {
	case errors.Is(err, voicevox.ResultGPUSupport):
		return fmt.Errorf("voicevox: GPU not available, check the CUDA download: %w", err)
	case err != nil:
		return fmt.Errorf("voicevox: unable to create synthesizer: %w", err)
	}

	return v.loadModel(filepath.Join(datadir, "models", "vvms"))
}

func (v *Voicevox) loadModel(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.vvm"))
	if err != nil {
		return err
	}

	var available []string
	for _, f := range files {
		model, err := voicevox.OpenVoiceModelFile(f)
		if err != nil {
			return fmt.Errorf("voicevox: unable to open voice model %s: %w", f, err)
		}

		metas, err := model.Metas()
		if err != nil {
			model.Delete()
			return fmt.Errorf("voicevox: unable to read voice model %s: %w", f, err)
		}

		for _, c := range metas {
			for _, s := range c.Styles {
				if s.ID != v.style {
					available = append(available, fmt.Sprintf("%d (%s %s)", s.ID, c.Name, s.Name))
					continue
				}

				err := v.syn.LoadVoiceModel(model, voicevox.DefaultLoadVoiceModelOptions())
				model.Delete()
				if err != nil {
					return fmt.Errorf("voicevox: unable to load voice model %s: %w", f, err)
				}
				return nil
			}
		}
		model.Delete()
	}

	return fmt.Errorf("voicevox: style %d not found in %s. available styles: %s", v.style, dir, strings.Join(available, ", "))
}

// Close frees the synthesizer.
func (v *Voicevox) Close() {
	v.syn.Delete()
	v.syn = 0
}

// Speech returns WAV audio for the Japanese text.
func (v *Voicevox) Speech(text string) ([]byte, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	if v.speedScale == nil && v.pitchScale == nil && v.intonationScale == nil && v.volumeScale == nil {
		return v.syn.TTS(text, v.style, voicevox.DefaultTTSOptions())
	}

	query, err := v.syn.CreateAudioQuery(text, v.style)
	if err != nil {
		return nil, err
	}

	var q map[string]any
	if err := json.Unmarshal([]byte(query), &q); err != nil {
		return nil, err
	}

	setScale(q, "speedScale", v.speedScale)
	setScale(q, "pitchScale", v.pitchScale)
	setScale(q, "intonationScale", v.intonationScale)
	setScale(q, "volumeScale", v.volumeScale)

	b, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	return v.syn.Synthesis(string(b), v.style, voicevox.DefaultSynthesisOptions())
}

func setScale(q map[string]any, key string, value *float64) {
	if value != nil {
		q[key] = *value
	}
}
