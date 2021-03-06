package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"encoding/json"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astichat/builder"
	main "github.com/asticode/go-astichat/server"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/xlog"
	"github.com/stretchr/testify/assert"
)

// Constants
const (
	// prv1 has passphrase "test"
	prv1String = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpQcm9jLVR5cGU6IDQsRU5DUllQVEVECkRFSy1JbmZvOiBBRVMtMjU2LUNCQyw3ZWNjMTQzYzkyNWI2ZWNmOTkxODkyOTc1OTdmM2RiNQoKc0hBVGpJZ1lUSi8zZ1BJTHFDR3BxaVEyYk9FbEd1cmlnYUZrdnRNZ29la3d5L29SYWFDRTBjN2hoeTdBYlNBTAprRm5qNkN4TWtZMElBRkRFRGt2S3J4VWpiNXR2Zjh2U21lZjBhQ2Z5U3pvUUFvWUV0aE9HY08wcGYxR3lISFN2CkFTU082anRsYUxieEExRU9BOTBCMEpFdHNUZjR3d0xvdGlZeG1UMGkwVU5wck5sMGtqWkdPTm1XM2xJZXhnTXgKNHkrcjdhY2ZGeGR1M2VzSDdtd04wMWVoVklMeGw3d0NmdHBFMnR0Z281RXhjMk9oSFY3NjZWek5pQTZXU1Bucwp0Y2JHdTFUWWRORGRrRmYyT2dhMXpxRDc3OTM5T2dkelk3RkdXSGcxdStnTVNTeXdMb2xEVlljaU5ubHJtWkp3CjlIdDRRdC9OYUFSalVzY3gzdDY4MjZTbklQckgyeklvVThkdStENFl3N0dvWnBCYnZNV3paV1ZNS1FvUWtVZ0IKUFdtNUN3bzZXelBzYmFRTUFYdUdRSllXOXNDdk95YzJFWngzNURncURIakJ6V1JyUmpORTh3UUY1SytxbUE5bAp1UktoQnFvQVNOQWdUeEVMUGxoT0lIa21oSVAvMDRhMkFpbmpCejV1TktWR1kyZW5rZ1VyMldtQnBnOGdRNTY4CjBPK1NvbWkrYkx2TWF0elVZNU96cGZXOTBzT2p3V3lRdE1hejFVZC9COVIyS2VMaFk2VlJ6WFEySUppeWNsWWUKR2Nza2xpRlJLTWNRb3NZcUNEODVHN2x2bVJYdy9CcTVkVSt1Um1nQzB4cXJDdGdlODdiUjhqdTM3ZXB0UmQ1bgpJYUdYcHBVbzZBeTA3UitIOUVpR0dpdXRnWlk4UUlHZUdyQ3UxUitLc2UrbTdSajJtZUtoK0hqVmIrNHJreXd4ClNSUHRmNnhRSlcvWmlVdVNPdDRyQnNCeXRKcWlOL3R6VW1oaUtZMk03ODZLQTIyQ0grUkNPQlVDWFYvUUV3dDQKaTh5T0RnRzg4UXpkanY0T08wT1RYck8yUEtST3Q0SFZsZGhiK21hdk1BNXN0WnZhMlNjeWw0R2ROWWZORjYxNwpvWERVRVk0M3AvSGFudkg3aC9vNmM2NjNwdnNSRzF0cjNwUU0rS0txRWNKMzB6eWcvNjFzaDRUNUNZeTkweUhNCk5zbFlVTjlHaWtpYkNCWW1nOVg4eUl2Slp4R29waUF5b0Nrb1k4bXdHQkxqSHpYNzJIdlZpYTZoMnZqRE4yaTAKM202VXc3UXkwZ0x1M1kwM3kvYkxzY3B6YXYrcWFGWDdxaVpvUWtRRS9IU213QU45K1IrYmlMTk4xdkRUdFVKZgp0M2d1ZlNEa3N2dnJVdmxBTTZVTUFxTmVRN3NwSnVTejBxTEd5emFGcnFsNVFXdmVDWFNCcitjZGFxaG5qVDgyCm16SlBBRVlZWVcvY3RQbmxUMnNRd29ZTEhUV2hhZC9sWG4rQ0c2ZXBTcFFnbllDOG10cUpvTW9pY29pQis3ejUKejdsMW8yd3dQdGhaRnQzNkE1VzhLUTc0d1JER21QdzFTeXprUDdPR1ZQU2orZ056KzIzTFhPU0JRRStPYnFlMQpGVm81ZUdWUzdmNTNmdW1Uc2YzUW1ENllnV1N0RFkzMHBaQm1yU2tiRVYyS240aUxheUxXdFlUdEUyYlVGL0E1Ck9wNTlsOUZLSXI3cHpPb1g2Q3ZrZlVUMTBNc0x6aTRWd3VEZTJwa2RRelE4cUNxOXdOMWlOYUR6bkh6eFpQRWEKSDRZUUpadlM4NGpRV3Q5QzF4enFqMWt5U0orV3RyMndFN2lQUm1WWDhIUkVQeUpzQU53MjdicHBHOUhpbFhGQgovam9JejBldm4yWkdpVCsxd0l4Z1lGbUt2TTBUNnp5Zm5DZTVGWmhlWnVxeFNHaERydHkrZzBtSGlzb3hSUDZ0CjFDT1BYanpSWkhqNGd6MzVJT0tvYUJNVkNIUCtyNU41VUthejk5ZGJYZWFvTlh0RHpJbE1zT1lJNDlTY3Y0UE0KSC9MY0pFZXl0d3lVZ0dCV1Nua0lKTHd2WVp6bGpEMkpiTTkrWlE3L3FuOUlqbjA4ZG9EVUhtUkovMExjUjM2dQpIVldDZ2xIaXpxbGRPRUkzODdDWklFWUxWZENtUVlnWUxHVnptNmplOUtlTDRNLzV2SkhyR2lTT2E5UEZIRG8xCmNlSjZXUGI3RG9qT1VNMlZJT0VsWUZoTUJ5RUVYM3cycVY4dytISTJXRW5hbXM1dGl6d0lCNnFYUXV0MWdCdkoKZVRQMEFobGdmNkg5YkVQcDBKNS9DdTBvZUtmMDZCalhaVW1vdjFGRStpTTdZR21OVnJrNUd3RVc3WVA1cGZaegpBeXJuS3BIb3U0RUk3cUhQZDEyN0g2MXR0YzQ0TVNtRmpodXZHNktrRXhoRkdFZWJPeWpManAxNHlNSWJ5eDlJCmNXN0NtNEZDN3EzY3dUZ3ZiSzhKYkdOaWNrR3BEdmdjenRJK1dvYUQ1ekVydHJNTFcrbXpaeWRlQXh3aFNzK3gKbVBKRmN3N3NrRS9qZjJnbUg4ZmtZdFVsaytIN2VhR01vb1lRaUNKNnRURWtwWTJudkp1TUFhU3M0MEx5V3MxKwpzc1pHZWxzcXBNRjZVYzVQeEcrZVI1a09QNkJjbVFSUml6MUIrV01OdXRGYjYrSjRQZFNjYmhZOEZZRVJYd3I4Ckc4Mm5qN1daOEt6b1piYzlJb1ZQOFNncTFKODhZVVhJUEVSZnRZeHpneWVDalRrV1ZyZi8rNEJKNVU0S3hXVkEKM0VlczEvSzJvUS9ZUXhyK2l1emc1UXFaNVVDelJ6SFVCRmM5UkN5VzN5eXdITTlKbXZNSXVCNXZYcG0yUm8ySgpOTkxEZ3NHQXI5UUtrbnZZTkFETVd3eGp5dm1hb3BETmViZXhoR3MzWHpnUWdrdGI5VjdReTNuc2IwUkM3MkNCCkpZQnZidmpjUEJ5YS80UU9OSnFNazdDZFIzVHNYUXdINUJGU2JiaHhFdjU0dlF2aldiZU5BVDd1dXhqK2lJYk4KZFRvZFR1QlhXY0V0Y3Y2ZVc0VzRhaGMzN3dJWndSQkhHblFZV2dyZEFEQkRVaytXWnBvTlBhZGJJaTlKUmU1VQo1S01zaDVweW9VTUNjQjQ3Zi9kQUVuNGkwWWlnWG5RcmdIcnJNSHY0aFJQYXVSR054SkwxdnNMd1AwemoxZ3ZNClRYQnY0aTRYbDRWdXhoWFlQUnJDWUFlL3hTNjhJcHhuRjlOZkZxbnd0VlRCd3l5aWs0MDJHWDVJTkFLYnB0UFYKeGtKN3l6dGpwdS8vdlhXNjZ5ZnVMcGVobGNkNm95OGFtczRzV01JRXFkNE1tTS9PbXFRTFNaaTdTZ01uOHY2Zgp4Y1F4ekxiUTNncm92dnROalVWc3dieVNUT0VCY0dkU0ZMY3AweERFTXRyYTNqTTBGajBsYWx1OXhWSHBXWVVMCkRHeEtRSEdsUUpHZ2MvU2dnT3lTTDFQVVk4SW5BRDlTNysraHNjR0lHMVRqdEJEUldIdzJWdG5XTU85WlBMWE8KeXc3ODQxN29EQ2dLN001NWZjWVV5b1dhNDJsNXRhWjhLVmg3by8rUHkwVnArSTRnQ2FValJEYVV3YmRMckRUZgpoeGFVK3Vid3YxaTdJNzlkRjY4Qzcxa1FkeHVIeUIxRnNXTUJiUUlVbDBQdHlXcDdNVjcxSzBEVGpRaXlOMWlBCm0xYVlUcVdud0x2M3lSdTEyWmxnNWdLOHQ5SGxKS1hNRDJ6RndxeHZSWWRoY01wZmlsNVBwMEVRcnBOcms1bFEKOEN0SmRlRmYzSUVzU200bm10N0ZHVStlKzR0NkRLcTB4Q3MwYW52bXJnUC9CazlyWEYwbVdlRHFMdkFscW1lQgpEY2ZJbFpjdmRhTnh3aWtHdGhjY2RtQkNlMmIyVTNsb09zc1VVNFRCdWZ2eEhqbVlGVVNuWnRPc1BRZTdQK0M2CnNrVXdrSDRUb1lJQWl3Wm8xTm1JbWNzVzRSR2xpTmRySmhaQWFFVGxPT2kybGQ2Rk9vRDZPeFB6RFpnRFNJcHgKeUs2Z21NZ29kWTFXKzA0V0tmLzNLWUtLZllnVHI5ZkZmWWpBcVVpdExZekdDM1d2MmwyZjZnRm56L3RkcndaTgotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
	prv2String = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKS1FJQkFBS0NBZ0VBMURmSm8rN1RMSTZBQ1VMdlJRcFBjS3J3U2FVVVBFVUpGSC9FVjVUbGZKeWhYRnJJCkQ2a1laKzhmVjRmVndZMEE2enRXSk1YRHhlSWNxMW8ySndNdDRYNFp4MjRYMDYvclNwY3RyYmNrWEZUNHNvcGYKMjRua1h5OGNVclRlaXBsYnQ4bmZ1eGlScXZhY2d1cTI3U0MvRWJybGtremYxWjNWMm15WGNVdlR0RTY1UjZLawpVdXU4VWZwaTFPWnl1QnZFUWJ2dzZPSjd5MXpTUXpsd2xMUDBUeWZ6aW0yMGVNSThyUXp3cmJ5ekpjQ2JLWEpGCml1ZFFXRDgrMC9JOWg5UTk5MXRWdHR2cU80VDFEUzZxN0xON0pPSFZMSDYxM0c4cklJeW5sNmxvU3cyMVViS1YKdkhMTThjTjR0a1lYZGZHUTEzRjREeDNJRVAvUGNES2lqMVlSTnJHU2ZnT084dVdSc2QzTkJiTFhEVThuNXJIRwo5UUphcEJEWDFlbVFWdkFsemxJdnJSb082V25XT0NLbGFFYUw4ZDVuZXNETjdrWXV2WVp6ME05am5KVVB0VUVyCkVFUUw2VzRocFB0di9YVXpaeXVqNWVFa010NmYvOXkrckpFOURXU2lzZnQvYVcwaFlSQWFONGtVS2VjMC93UTQKWjUzc1d2NnQ3SWpyRkFKN25lSDNMQkFsWnhkS2FlNy9KSjAzdjBDWW1CM1c2aTBBSmIreFZ6dlJvSVhLYWVveQpRdWdyeGF1RmZOSXgwVGxaNzd4bTRrTStaSnhRSTM5Q3h1RXN0aEM3TFlLM1N1akJFMHZEQnkrOVcwTXBEbGhwCkN1UXhmUWU3MW5vdTFqZnhrdFBuRmcxQ2hrYzFEUG5PMVBZRm1Rc0pENHN0RnJIRXZPUTVTcmMrMTFNQ0F3RUEKQVFLQ0FnRUFnb0c0V0Q4ejBLL2xuMHh4ZHFUTGk3OGp2RFp2eGt5eU05QUsvODFLZjZLWFBRTjdDdDV6YXQ5TQpCL2s2QkNoaGkwZlhSdy96d0VxNFZNeEtoeDFXWnRpMG84ZFprYzRheGFsSTV3NjhwcWQrdGRXUTg2TE9OWmIwCk5ReVQydXBLMURDcWpSV2o1MTUzaTY4cVJaT2d6UmVCdk1IWDJUZVNYeHZ1MmpiR2Y1ajJLazZqL1hhSlBtVGIKeUkvYnRzc2ttMFFuK0Ivbi8zMGF0VXFxcUZndWcwdFBZeTdxRUdWckNRVHZNZmpjdHZmR3MrdFpSdjNQbENWNAp6c0NuQkZRS3M0YVFwTDZEUW8wV1lqL3p6MUxsQlI1NGlUOTNPWk9JRXlGTW8yRUVDVHZwNk04SmRIV3BBWGl6ClVHeTBXc3p1eFA4NzFSZjhoQys4OHdQQW9xTk1PNVlIVmp0RUZOdHR1TS9qcVhMT1VlMjBOZ280ME1pK3Y3YjQKaXBwWWMxbjU1M0I5LzBHQXQ0cEcvbDhETDhJMWwzNG92RzhkUG9JZCt0WGNTMEVVRFZoL1JWS0wxUkt6SWUySQpEc3ZuNVIvQzJnNEN1UjN6K0NzK1hKOS9kZER3YkZiMWNYUnQvU0JpdTVWSlk0aGZ1THQ2RXp3czlQWGNiYmNmCm1NYlhHQ0pKWnhRSkNHTGJ4d28wYTAxT2QzMDBJSUNxWFg0ZUxUc2VxZ0xaNS9nOEdubG9SMDhNNzFFREJNVCsKWEx5NWlyOFR4UnJIb3hvdWE5WE5wdE04cjFCdC9PTzNxUDlIVGRXd2I0UEw4TWkxUXJBOXdTcXVsWmV3bDZSbAp1bllMVlBGSG4rdHRvKytUUTd5Y0kwa1c5azdqcXpINkJ5dVVCRm1Zb1Z3WlA5dlFVeUVDZ2dFQkFObWlWVnNMCjFZekUxdkJNTThORVhoeWFTTlJicmc1bXgxMzUzMGRQdGFncDE1QXpFMlU4eTh5b0QrNGFES29hNFhvdjRsK0gKUi9PUHBwcHNwdnBFRUg5QnFwblE1QnFrRW91VTFFY294RHFrVUR0TEIxYVl1Y3Z4L0pnNit2VThqaWZDay8veQpsN08yR3BaRVo5Z2toVi93RjJweXZGaEVWK21CK2RBaVpDaUx2TzR1MFFaQ08wbnZsNWdzdm0xeU9rWHo2RXpTCmNpSmc4YlZXOUg5MGoxQ0RneldXWGV3U1drOXNOU0VCMHpSdGRidVFuV3ZIMVlDN3VWbFpGUmdIOWh4Y05QWFQKeUw0V3NLckk0eW40WnZiTGlQemhqRHpxZ1BFN3pxeFdoZzM3c3pGN1BkS1BXQkRxZEtFZWxCekZ5QXhXdDhJUgprczRDYXFkRHFzSmNUM0VDZ2dFQkFQbWhCb0FXTjVieGdRZ3cyUUpFbnhPRVdkTysvUlZVRHZnam1oR0pnZEZOCjB3aEJQWkFPWlJEOXBtS3BiM3ZQVHcrTks1KzhjUUpTL09mYW5lNE5iQzY1eVY3R1RaVDhkcEV1dVFNTXhmZzUKOHpxOFNDWlJZTG1QTnB4ZDU0aksxYVBhaEx0RUpNK1lLR3pPd3dpS0tHL00xOXZ1b2FnUVJrbmFLeEt0TnhZOQpDR0hLUys4OWZUdkNvOVVYcXZLbVdrVXoyM1VMVlVUMjhPZHpNb0ZLWWI4cjdUSHVHc1hhSVc0YkFteWQrbjhSCk03ZFhTZ1VEK0kzUEZ1QW5PbUE3NGFDMmtsRXhHSlJPOHRjc3d2U01rb3ZFV3ZmNHJ2N3gzNnRzam5jcGdNVmYKNzZKdnovQlB6YVVxRVZrd1Q2WkRTbE1TMjk2QmU0SDFqK3FwSER5MStRTUNnZ0VBQm5WNzVRVFg1S0tlNG1qUQpqSFlGK1FGWE1mNDZqekRicjkxUGxCVTRoZklmOUthZlo3ekRLNks5UGtyRm4zTEd5RktOZkZwT2Qxc0hEY1ExCnZHMnlzNlFtUlFSZkVLOVg0WTZjTWpSeWhtOEQ5bzZHZkRweUlTeGdXOEE4WEhUY255OTJKdjF6SlNFOWJzSSsKOXJvMnZ4OG9BcisrK1R1bUJFY1lPK1laWk42b3o4VFI3VWFmN2RUUGdmT3AveU9KdVRQQTdDNit0bWg4SSs2ZAp1UDZqbGpjZytNRXFybXZwQkR4bzR6N2puc1cwM2Nrdjh0ZnViVENsRXBMRFBvQlYvSWQ4QnVPdkxIME41ek9wCkVrRE9CWHNLNkw1azVCWHRsN3MzcWdPelhNemdoNUpweGtyOHlSdThORi9zODJHblN6NXptNjNiMW9OWjJQYjQKSldhSjhRS0NBUUVBMURPMVVlY1JCR2h4OXFPSHBpenRTV3NGN1VGNjVNbWJIQWN2cmw3RkUwYmo4UzE4aHR3bAp3QWJQalNsWmt0Y003enhqYkJ1RnVhTVFTSXdJR3RnZ01heFBhUmlMMU8yMFlRQmwyQmpncFgybHJUVm00K1BqCnBIb0F3M1gwSDgzRlJNKzZhM0tuRkMzVmw0RkFQQjh0OXJRY2YySmcyM3hTTSsrWkUveFpTcmRCUzlmckt3bUwKVHVUTDNwYUxCRkN6aGdacU5Sb1lOMUx3UU9BbGU5RGVQT083YytsanF2TWQzZnBwMmltRlNzVTF3Rkljb3h6WQpDcnlUUnFNeU5hSlIwQXZEWCsrclpFK2trWlFFZWx5Ukt1MFZJNXlzTGg4d3N3bktKYlFMT2oydWVOZ3gzS2dLCk9iQUVKVnd6S1RRa2wyLzlwaTFONzVEdThWMG1tdGxhUHdLQ0FRQThsRFV4TDA1RXJ2VFJuQTcxSWppNjZSdTgKV21IajBrTW5sQVF0aE1hV04rYXNucE03WU9JUXFNbHBoM0pMUXMweHExK1h6RjFXcDRlaGhiVHJvUjczNkZ2MQpmSlBLemVoZGFwK290b0hWVUpidVdaQzZLNnRSZDlPWWdNNmNGOXNXd0xDc1FuOUVjbEhZaW5CNnFLbFJWeUVFCllMNjRKRW1SY0o2ZFNBWnNaMm9hUlNZdFdwbXQ4MEpqUVdPeE0rK2JTN0FOdWlkVTdseHdaOWhrSXI4NUVHb2oKTmQwL0JDTThxbUltWVNrKytyRUlmKzBXbU5OaUFhSWdmQ09XNzh6aU9HZzU4L1JjL2lyRFBTd3pKbVZwVktORQptZHJmVUh4bWtxbXh5L3piK211K3lpdFRlbWNpU0ZLcXRzSkMyODh5WFdPbE1zblRlbjc5eVBRQXA0UUIKLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0K"
)

// TODO Fix tests
func TestHandleDownloadPOST(t *testing.T) {
	// Init
	var l = xlog.NopLogger
	var s = astichat.NewMockedStorage()
	var rw = httptest.NewRecorder()
	var r = &http.Request{}
	r = r.WithContext(main.NewContextWithBuilder(r.Context(), &builder.Builder{}))
	r = r.WithContext(main.NewContextWithLogger(r.Context(), l))
	r = r.WithContext(main.NewContextWithStorage(r.Context(), s))
	var prv1 = astichat.PrivateKey{}
	prv1.SetPassphrase("test")
	var err = prv1.UnmarshalText([]byte(prv1String))
	assert.NoError(t, err)
	var pub1 *astichat.PublicKey
	pub1, err = prv1.PublicKey()
	assert.NoError(t, err)
	var prv2 = astichat.PrivateKey{}
	err = prv2.UnmarshalText([]byte(prv2String))
	assert.NoError(t, err)
	var pub2 *astichat.PublicKey
	pub2, err = prv2.PublicKey()
	assert.NoError(t, err)
	var count int
	main.AstichatNewPrivateKey = func(passphrase string) (*astichat.PrivateKey, error) {
		count++
		if count == 1 {
			return &prv1, nil
		}
		return &prv2, nil
	}
	var ios, iusername string
	var iprvClient *astichat.PrivateKey
	var ipubServer *astichat.PublicKey
	main.BuilderBuild = func(b *builder.Builder, os, username string, prvClient *astichat.PrivateKey, pubServer *astichat.PublicKey) (string, error) {
		ios = os
		iusername = username
		iprvClient = prvClient
		ipubServer = pubServer
		return "/test", nil
	}
	var removed []string
	main.OSRemove = func(path string) error {
		removed = append(removed, path)
		return nil
	}
	main.IOUtilReadFile = func(path string) ([]byte, error) {
		return []byte("binary"), nil
	}

	// Empty username
	main.HandleDownloadPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "{\"error\":\"Please enter a username\"}\n", rw.Body.String())

	// Username is not unique
	rw = httptest.NewRecorder()
	s.ChattererCreate("bob", &astichat.PublicKey{}, &astichat.PrivateKey{})
	r.Form.Set("username", "bob")
	main.HandleDownloadPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "{\"error\":\"Username is already used\"}\n", rw.Body.String())
	s.ChattererDeleteByUsername("bob")

	// Password is empty
	rw = httptest.NewRecorder()
	r.Form.Set("username", "bob")
	main.HandleDownloadPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "{\"error\":\"Please enter a password\"}\n", rw.Body.String())

	// OS is invalid
	rw = httptest.NewRecorder()
	r.Form.Set("password", "test")
	r.Form.Set("os", "invalid")
	main.HandleDownloadPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "{\"error\":\"Invalid OS\"}\n", rw.Body.String())

	// Success
	rw = httptest.NewRecorder()
	r.Form.Set("os", builder.OSLinux)
	main.HandleDownloadPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, builder.OSLinux, ios)
	assert.Equal(t, "bob", iusername)
	assert.Equal(t, prv1.String(), iprvClient.String())
	assert.Equal(t, pub2.String(), ipubServer.String())
	assert.Equal(t, []string{"/test"}, removed)
	var c astichat.Chatterer
	c, err = s.ChattererFetchByUsername("bob")
	assert.NoError(t, err)
	assert.Equal(t, prv2.String(), c.ServerPrivateKey.String())
	assert.Equal(t, pub1.String(), c.ClientPublicKey.String())
	assert.Equal(t, "binary", rw.Body.String())
	assert.Equal(t, "6", rw.Header().Get("Content-Length"))
}

func TestHandleNowGET(t *testing.T) {
	// Init
	var rw = httptest.NewRecorder()
	main.Now = func() time.Time {
		return time.Unix(100, 0)
	}

	// Assert
	main.HandleNowGET(rw, &http.Request{}, httprouter.Params{})
	assert.Equal(t, http.StatusOK, rw.Code)
	var now time.Time
	var err = json.Unmarshal(rw.Body.Bytes(), &now)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(100, 0), now)
}

func TestHandleTokenPOST(t *testing.T) {
	// Init
	var l = xlog.NopLogger
	var s = astichat.NewMockedStorage()
	main.GenerateToken = func() string {
		return "new id"
	}
	main.Now = func() time.Time {
		return time.Unix(100, 0)
	}
	rw := httptest.NewRecorder()
	r := &http.Request{}
	r = r.WithContext(main.NewContextWithLogger(r.Context(), l))
	r = r.WithContext(main.NewContextWithStorage(r.Context(), s))

	// Empty username
	main.HandleTokenPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "{\"error\":\"Please enter a username\"}\n", rw.Body.String())

	// Username doesn't exist
	r.Form.Set("username", "bob")
	rw = httptest.NewRecorder()
	main.HandleTokenPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusOK, rw.Code)

	// Username exists
	s.ChattererCreate("bob", &astichat.PublicKey{}, &astichat.PrivateKey{})
	rw = httptest.NewRecorder()
	main.HandleTokenPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusOK, rw.Code)
	c, err := s.ChattererFetchByUsername("bob")
	assert.NoError(t, err)
	assert.Equal(t, "new id", c.Token)
	assert.Equal(t, time.Unix(100, 0), c.TokenAt)
}
