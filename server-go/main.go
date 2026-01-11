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

	//Diretório base para salvar os arquivos temporários
	// videoFile := filepath.Join("downloads", "temp", title+".mp4")
	// audioFile := filepath.Join("downloads", "temp", title+".mp3")
	
	// Arquivos temporários
	videoFile := title + ".mp4"
	audioFile := title + ".mp3"

	// Baixar vídeo, o exec command executa comandos do sistema operacional
	cmdDownload := exec.Command("yt-dlp", "-f", "mp4", "-o", videoFile, url) //baixa o vídeo no formato mp4, executa o comando yt-dlp e salva como video.mp4
	//-o é o nome do arquivo de saida
	//-f é o formato que o arquivo vai sair
	if err := cmdDownload.Run(); err != nil {
		return fmt.Errorf("erro ao baixar vídeo: %v", err)
	}

	// Converter para MP3
	cmdConvert := exec.Command("ffmpeg", "-y", "-i", videoFile, audioFile) //converte o vídeo baixado para mp3 usando o ffmpeg
	// -y sobrescreve arquivos
	// -i nome do arquivo de entrada
	if err := cmdConvert.Run(); err != nil {
		return fmt.Errorf("erro ao converter áudio: %v", err)
	}

	//remover video temporário
	if err := os.Remove(videoFile); err != nil {
		log.Printf("aviso: não foi possível remover o arquivo de vídeo temporário: %v", err)
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
