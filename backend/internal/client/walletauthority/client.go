package walletauthority

import (
	"dd-prediction-api/config"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/apperrors"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"

	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
)

type IWalletAuthorityClient interface {
	R() *resty.Request
	GetTxFromContract(reqBody map[string]any, endpoint string) (*solana.Transaction, error)
	GetSignatureFromContract(reqBody map[string]any, endpoint string) (string, error)
	GetInitTxFromContract(reqBody map[string]any, endpoint string) (*solana.Transaction, string, error)
}

type Client struct {
	client *resty.Client
	apiKey string
}

func (c *Client) R() *resty.Request {
	return c.client.R().SetHeaders(map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
		"X-API-KEY":    c.apiKey,
	})
}

const baseURL = "https://wallet-authority.overdone.it"

type Option func(*Client)

func NewClient(c *config.Config) *Client {
	client := &Client{
		client: resty.New().SetBaseURL(baseURL),
		apiKey: c.Client.WalletAuthorityAPIKey,
	}

	return client
}

func (c *Client) GetTxFromContract(
	reqBody map[string]any,
	endpoint string,
) (*solana.Transaction, error) {
	const url = "/main/"

	resp, err := c.R().
		SetBody(reqBody).
		SetHeader("Content-Type", "application/json").
		Post(url + endpoint)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to get transaction from wallet authority", err)
	}

	tx, err := extractTxFromResp(resp)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (c *Client) GetSignatureFromContract(
	reqBody map[string]any,
	endpoint string,
) (string, error) {
	const url = "/main/"

	resp, err := c.R().
		SetBody(reqBody).
		SetHeader("Content-Type", "application/json").
		Post(url + endpoint)
	if err != nil {
		return "", apperrors.ServiceUnavailable("failed to get signature from wallet authority", err)
	}

	if resp == nil || !resp.IsSuccess() {
		return "", apperrors.ServiceUnavailable("failed to get signature from wallet authority: resp is nil")
	}

	var respData model.WASignatureResp
	if err := json.Unmarshal(resp.Body(), &respData); err != nil {
		return "", apperrors.ServiceUnavailable("failed to unmarshal signature", err)
	}

	return respData.Signature, nil
}

func (c *Client) GetInitTxFromContract(
	reqBody map[string]any,
	endpoint string,
) (*solana.Transaction, string, error) {
	const url = "/main/"

	resp, err := c.R().
		SetBody(reqBody).
		SetHeader("Content-Type", "application/json").
		Post(url + endpoint)
	if err != nil {
		return nil, "", apperrors.ServiceUnavailable("failed to get transaction from wallet authority", err)
	}

	if resp == nil || !resp.IsSuccess() {
		return nil, "", apperrors.ServiceUnavailable("failed to get transaction from wallet authority: resp is nil")
	}

	var respData model.WAInitTxResp
	if err := json.Unmarshal(resp.Body(), &respData); err != nil {
		return nil, "", apperrors.ServiceUnavailable("failed to unmarshal raw transaction", err)
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(respData.RawTx))
	if err != nil {
		return nil, "", apperrors.ServiceUnavailable("failed to decode a transaction", err)
	}

	return tx, respData.RoomTokenPDA, nil
}

func (c *Client) GetPdaInfo(
	reqBody map[string]any,
) (*model.WAPdaInfo, error) {
	const url = "/main/get-pda"

	resp, err := c.R().
		SetBody(reqBody).
		SetHeader("Content-Type", "application/json").
		Post(url)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to get pda info from wallet authority", err)
	}

	if resp == nil || !resp.IsSuccess() {
		return nil, apperrors.ServiceUnavailable("failed to get pda info from wallet authority: resp is nil")
	}

	var respData model.WAPdaInfo
	if err := json.Unmarshal(resp.Body(), &respData); err != nil {
		return nil, apperrors.ServiceUnavailable("failed to unmarshal raw pda info", err)
	}

	return &respData, nil
}

func extractTxFromResp(resp *resty.Response) (*solana.Transaction, error) {
	if resp == nil || !resp.IsSuccess() {
		return nil, apperrors.ServiceUnavailable("failed to get transaction from wallet authority: resp is nil")
	}

	var respData model.WATxResp
	if err := json.Unmarshal(resp.Body(), &respData); err != nil {
		return nil, apperrors.ServiceUnavailable("failed to unmarshal raw transaction", err)
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(respData.RawTx))
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to decode a transaction", err)
	}

	return tx, nil
}
