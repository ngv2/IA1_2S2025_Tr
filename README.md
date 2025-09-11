# IA1_2S2025_Tr

### Ejecutar backend

```bash
cd ./backend/
go mod tidy
go run .
```

### Ejecutar frontend
```bash
cd ./frontend/
npm i
npm run dev
```

### Ejecutar RPA (visualizacion)

```bash
cd ./rpa/
go build -o rpa_prolog_loader.exe main.go 
./rpa_prolog_loader.exe C:\path
```

### Ejecutar RPA (Guardar contenido)
```bash
cd ./rpa/
go build -o rpa_prolog_loader.exe main.go 
./rpa_prolog_loader.exe C:\path_origen C:\path_destino
```