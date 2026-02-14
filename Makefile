.PHONY: proto run-gateway run-user run-content run-message run-chatbot

proto:
	protoc --proto_path=api \
	       --go_out=paths=source_relative:api \
	       --go-grpc_out=paths=source_relative:api \
	       api/user/v1/user.proto \
	       api/content/v1/content.proto \
	       api/message/v1/message.proto \
	       api/chatbot/v1/chatbot.proto
