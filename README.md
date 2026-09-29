# sayanything

Go package wrappper for Text To Speech (TTS).

Supports the following local TTS engines:

- [Piper TTS](https://github.com/OHF-Voice/piper1-gpl)
- [Software Automatic Mouth (SAM)](https://github.com/deadprogram/sam)
- [VOICEVOX](https://github.com/hybridgroup/voicevox) for Japanese

It also supports the following cloud based TTS engines:

- [Google Cloud TTS](https://cloud.google.com/text-to-speech) 
    NOTE: has not been used or tested in quite a while.


## How to build

```
go install ./cmd/sayanything
```

## How to run

### Piper

```
sayanything --engine piper --voice hfc_female-medium --lang en_US --data ~/voices/ "hello friends"
```

### SAM

```
sayanything --engine sam "hello friends"
```

### VOICEVOX

To install libffi and the VOICEVOX runtime, follow the steps in [hybridgroup/voicevox](https://github.com/hybridgroup/voicevox#installation). Then pass the runtime directory with `--data` and a style id with `--voice`.

```
sayanything --engine voicevox --voice 3 --data ~/voicevox_core "こんにちは、世界"
```

You can adjust the voice with `--speed-scale`, `--pitch-scale`, `--intonation-scale` and `--volume-scale`.

```
sayanything --engine voicevox --voice 3 --data ~/voicevox_core --speed-scale 1.5 --pitch-scale 0.05 "こんにちは"
```

To use CUDA, add `--gpu`. This needs the CUDA runtime from the VOICEVOX downloader, which is described in [hybridgroup/voicevox](https://github.com/hybridgroup/voicevox#gpu).

```
sayanything --engine voicevox --voice 3 --data ~/voicevox_core --gpu "こんにちは"
```

If you pass a style id that isn't installed, the error lists the styles that are available. The VOICEVOX terms require a credit for each character you use, for example "VOICEVOX:ずんだもん".

### Google Cloud TTS

```
sayanything -k="/path/to/key.json" -l="es-ES" -voice="es-ES-Neural2-E" "¡Hola amigo! ¿Cómo estás?"
```
