package main

import (
	"fmt" //usado para formatar mensagens de erro
	"io" //usado para detectar o final do arquivo
	"log" //usado para logs no terminal
	"net" // usado para abrir uma porta tcp
	"os" // usado para manipulação de arquivos
	"os/exec" // permite executar comandos do sistema operacional (yt-dlp e ffmpeg)

	//"path/filepath" //caso queira trocar o caminho de download dos arquivos temporários
	pb "server-go/converter" // importa o pacote gerado pelo compilador protobuf
	"strings" // usado para manipulação de strings
	"google.golang.org/grpc" //biblioteca principal do servidor gRPC
)

type server struct { // implementa a interface gerada pelo .proto
	pb.UnimplementedVideoConverterServer //gerado automaticamente pelo compilador protobuf
}

// Método gRPC que vai ser chamado pelo cliente, recebe a URL do vídeo (req) e retorna o áudio em streaming (stream)
func (s *server) ConvertVideoToAudio(req *pb.VideoRequest, stream pb.VideoConverter_ConvertVideoToAudioServer) error {
	url := req.GetYoutubeUrl() //extrai a URL do vídeo da requisição
	log.Println("Recebido:", url)

	cmdTitle := exec.Command("yt-dlp", "--print", "title", url)
	output, err := cmdTitle.Output()
	if err != nil {
		return err
	}

	title := strings.TrimSpace(string(output))
	log.Println("Título do vídeo:", title)

	// Arquivo temporário de áudio
	audioFile := title + ".mp3"

	// Baixar áudio diretamente com yt-dlp (evita problemas com formatos de vídeo bloqueados)
	// -f bestaudio: seleciona o melhor formato de áudio disponível
	// --extract-audio: extrai apenas o áudio
	// --audio-format mp3: converte para MP3
	// -o: nome do arquivo de saída
	cmdDownload := exec.Command("yt-dlp", 
		"-f", "bestaudio[ext=m4a]/bestaudio",
		"--extract-audio",
		"--audio-format", "mp3",
		"-o", audioFile,
		url)
	
	// Captura a saída de erro para melhor diagnóstico
	cmdDownload.Stderr = os.Stderr
	cmdDownload.Stdout = os.Stdout
	
	if err := cmdDownload.Run(); err != nil {
		return fmt.Errorf("erro ao baixar/converter áudio: %v", err)
	}

	//  Abrir MP3
	file, err := os.Open(audioFile)
	if err != nil {
		return err
	}
	defer file.Close() // garante que o arquivo será fechado após a função terminar

	//  Buffer para leitura. evita que o arquivo seja carregado todo na memória
	buffer := make([]byte, 1024)

	//  Enviar em streaming, lê os arquivos em pedaços e envia cada pedaço para o cliente
	for {
		n, err := file.Read(buffer) //lê até 1024 bytes do arquivo e armazena em buffer
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		chunk := &pb.AudioChunk{ // cria uma mensagem protobuf AudioChunk com os dados lidos
			Data: buffer[:n],
		}

		if err := stream.Send(chunk); err != nil { // envia o pedaço para o cliente via stream
			return err
		}
	}

	log.Println("Envio finalizado")
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterVideoConverterServer(grpcServer, &server{}) //associa o servidor gRPC com a implementação do serviço

	log.Println("Servidor gRPC rodando na porta 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
