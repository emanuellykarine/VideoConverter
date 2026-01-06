# Video Converter gRPC(Youtube -> MP3)
Projeto utilizando gRPC que é um framework de código aberto para comunicação entre sistemas. Permite que um cliente chame métodos diretamente em um servidor remoto como se fosse uma função local.
Protobuf (Protocol Buffers) é um método eficiente e neutro de serialização de dados estruturados sando um formato binário compacto, ideal para comunicação entre microserviços via gRPC

## Arquitetura
- Cliente: Python
- Servidor: Go

## Dependências
- Go
    golang-go
    Linguagem utilizada para implementar o servidor gRPC.

- Protobuf / gRPC
    protobuf-compiler
    Compilador Protobuf (protoc), utilizado para gerar os códigos gRPC no Go e no Python.

    grpcio
    Biblioteca gRPC para cliente Python, responsável pela comunicação com o servidor.

    grpcio-tools
    Fornece o compilador Protobuf para Python (grpc_tools.protoc), usado para gerar os arquivos *_pb2.py.

- Processamento de vídeo e áudio
    yt-dlp
    Ferramenta responsável por baixar o vídeo do YouTube a partir da URL fornecida.

    ffmpeg
    Ferramenta para extração e conversão de áudio, utilizada para gerar o arquivo MP3 a partir do vídeo.

## Passo a Passo

- Instala plugins do Go
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
```

- Adiciona o Go ao PATH: 
```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

- Inicia criando o contrato gRPC (.proto) que define o contrato de comunicação. O cliente chama o método remoto ConvertVideoToAudio enviando uma URL, e o servidor responde com um fluxo de dados binários representando o áudio.

- O código gRPC é gerado automaticamente, dentro da pasta proto roda os seguintes comandos:
```bash
protoc \
  --go_out=../server-go \
  --go-grpc_out=../server-go \
  converter.proto
```
Isso gera:
structs
interface do serviço
código de serialização

- Código do cliente python
```bash
python3 -m grpc_tools.protoc \
  -I. \
  --python_out=../client-python \
  --grpc_python_out=../client-python \
  converter.proto
```

- Criar servidor gRPC em Go
```bash
cd ~/Documentos/VideoConverter/server-go

go mod init server-go

go get google.golang.org/grpc@v1.56.0
go get google.golang.org/protobuf@v1.28.1
```

## Executar o projeto
- Abre em dois terminais diferentes, um do servidor Go e o outro do cliente Python.
```bash
cd server-go
go run main.go
```

```bash
cd client-python
python3 client.py
```

