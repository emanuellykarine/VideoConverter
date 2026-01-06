import grpc
import converter_pb2
import converter_pb2_grpc


def main():
    # Conecta ao servidor gRPC
    channel = grpc.insecure_channel("localhost:50051") #abre um canal gRPC inseguro na porta 50051 que é padrão
    stub = converter_pb2_grpc.VideoConverterStub(channel) #cliente remoto que possui os métodos definidos no serviço VideoConverter .proto

    # URL do vídeo
    youtube_url = input("Digite a URL do YouTube: ")

    request = converter_pb2.VideoRequest( # cria a mensagem protobuf VideoRequest com a URL do vídeo fornecida pelo usuário
        youtube_url=youtube_url
    )

    # Arquivo de saída
    with open("audio.mp3", "wb") as f:
        print("Recebendo áudio...")

        # Chamada RPC com streaming
        for chunk in stub.ConvertVideoToAudio(request): # o método retorna um iterador que produz objetos VideoChunk, cada iteração é um pedaço do arquivo de áudio convertido
            f.write(chunk.data)

    print("Download concluído! Arquivo salvo como audio.mp3")


if __name__ == "__main__":
    main()
