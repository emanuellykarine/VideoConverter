package main

import (
	"fmt" //usado para formatar mensagens de erro
	"io" //usado para detectar o final do arquivo
	"log" //usado para logs no terminal
	"net" // usado para abrir uma porta tcp
	"os" // usado para manipulação de arquivos
	"os/exec" // permite executar comandos do sistema operacional (yt-dlp e ffmpeg)

	pb "server-go/converter" // importa o pacote gerado pelo compilador protobuf

	"google.golang.org/grpc" //biblioteca principal do servidor gRPC
)

type server struct { // implementa a interface gerada pelo .proto
	pb.UnimplementedVideoConverterServer //gerado automaticamente pelo compilador protobuf
}

// Método gRPC que vai ser chamado pelo cliente, recebe a URL do vídeo (req) e retorna o áudio em streaming (stream)
func (s *server) ConvertVideoToAudio(req *pb.VideoRequest, stream pb.VideoConverter_ConvertVideoToAudioServer) error {
	url := req.GetYoutubeUrl() //extrai a URL do vídeo da requisição
	log.Println("Recebido:", url)

	// Arquivos temporários
	videoFile := "video.mp4"
	audioFile := "audio.mp3"

	// Baixar vídeo, o exec command executa comandos do sistema operacional
	cmdDownload := exec.Command("yt-dlp", "-f", "mp4", "-o", videoFile, url) //baixa o vídeo no formato mp4, executa o comando yt-dlp e salva como video.mp4
	if err := cmdDownload.Run(); err != nil {
		return fmt.Errorf("erro ao baixar vídeo: %v", err)
	}

	// Converter para MP3
	cmdConvert := exec.Command("ffmpeg", "-y", "-i", videoFile, audioFile) //converte o vídeo baixado para mp3 usando o ffmpeg
	if err := cmdConvert.Run(); err != nil {
		return fmt.Errorf("erro ao converter áudio: %v", err)
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
