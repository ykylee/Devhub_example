package webhook

// init 은 package import 시 자동으로 모든 adapter 를 registry 에 등록.
// IngestIntegrationProviderWebhook handler 가 GetAdapterForProviderType 호출 시
// 4종 adapter (Gitea/Jira/Generic + 추후 1종) 자동 dispatch.
//
// 본 init() 은 IngestIntegrationProviderWebhook handler 가 본 package 에 의존하는
// 모든 backend binary 에서 자동 실행 — 별도의 main.go 변경 불요.
func init() {
	RegisterAdapter(NewGiteaWebhookAdapter())
	RegisterAdapter(NewJiraWebhookAdapter())
	RegisterAdapter(NewGenericWebhookAdapter())
}
