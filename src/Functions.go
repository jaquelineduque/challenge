package main

/*
import (
	//"strconv"
	"regexp"
)
*/
func FInsertClient(client Client) (Client, error, CustomError) {
	return InsertClientBD(client)
}

func ConsultClients() Clients {
	return ConsultAllClientsDB()
}

func ConsultEspecificClient(id int64) Clients {
	return ConsultClientDB(id)
}

func ValidCardInfo(cardInfos CardInfos) (bool, string) {
	var msg string
	if cardInfos.HolderName == "" {
		msg = "Holder name should be especified."
		return false, msg
	}
	if cardInfos.Number == "" {
		msg = "Card number should be especified."
		return false, msg
	}
	if cardInfos.ExpirationDate == "" {
		msg = "Expiration date should be especified."
		return false, msg
	}
	if cardInfos.Cvv == "" {
		msg = "CVV should be especified."
		return false, msg
	}

	if !ValidCreditCard(cardInfos.Number) {
		msg = "The card number isn't valid."
		return false, msg
	}

	return true, ""
}

func FormatClientConsult(cl Clients, state int, msg string) ReturnClients {
	var retClient ReturnClients
	if cl == nil {
		retClient.Return.State = 0
		retClient.Return.Message = "Client(s) not found."
	} else {
		retClient.Return.State = state
		retClient.Return.Message = msg
		retClient.Clients = cl
	}
	return retClient
}

func BoletoPayment() string {
	//Mocked value
	const boletoNumber = "23790504004199033014836008109203478470000019900"
	return boletoNumber
}

func CardPayment() bool {
	//Mocked value
	const successful = true
	return successful
}

func validPaymentType(paymentType int) bool {
	return (paymentType == 1 || paymentType == 2)
}

func PaymentMethod(payInfo2 Payment) PaymentReturn {

	payInfo := payInfo2
	var payReturn = PaymentReturn{}

	//Validates general informations
	if (payInfo.PaymentInfo.PaymentType != 1) && (payInfo.PaymentInfo.PaymentType != 2) {
		payReturn.Return.State = 5001
		payReturn.Return.Message = "Please, check payment type."
		payReturn.Return.TechnicalMessage = "Invalid payment type."
		return payReturn
	}
	if payInfo.PaymentInfo.Amount < 1 {
		payReturn.Return.State = 5002
		payReturn.Return.Message = "Please, inform amount of payment."
		payReturn.Return.TechnicalMessage = "Amount not informed."
		return payReturn
	}
	if payInfo.Client.Id < 1 {
		payReturn.Return.State = 5003
		payReturn.Return.Message = "Please, inform client."
		payReturn.Return.TechnicalMessage = "Client not informed."
		return payReturn
	}
	if payInfo.Buyer.Name == "" {
		payReturn.Return.State = 5004
		payReturn.Return.Message = "Please, inform buyer name."
		payReturn.Return.TechnicalMessage = "Buyer name not informed."
		return payReturn
	}
	if payInfo.Buyer.Email == "" {
		payReturn.Return.State = 5005
		payReturn.Return.Message = "Please, inform buyer e-mail."
		payReturn.Return.TechnicalMessage = "Buyer e-mail not informed."
		return payReturn
	}
	if payInfo.Buyer.CpfCnpj == "" {
		payReturn.Return.State = 5006
		payReturn.Return.Message = "Please, inform buyer CPF or CNPJ."
		payReturn.Return.TechnicalMessage = "Buyer cpf/cnpj not informed."
		return payReturn
	}

	//Check if client exists
	if !(ClientRegistered(payInfo.Client.Id)) {
		payReturn.Return.State = 0
		payReturn.Return.Message = "We've coundn't found this client."
		payReturn.Return.TechnicalMessage = "Client not found."
		return payReturn
	}

	//var retBuyer = Buyer{}
	var retBuyer = Buyer{}
	//var customError = CustomError{}
	//Check if buyer exists. If doesn't exists, save
	if BuyerRegistered(payInfo.Buyer.CpfCnpj) {
		localRetBuyer, err := BuyerInfo(payInfo.Buyer.CpfCnpj)

		if err != nil {
			payReturn.Return.State = 0
			payReturn.Return.TechnicalMessage = err.Error()
			return payReturn
		}

		retBuyer = localRetBuyer
	} else {

		localRetBuyer, err, customError := SaveBuyer(payInfo.Buyer)
		if err != nil {
			//avoid empty technical message
			tmessage := customError.TechnicalMessage
			if tmessage != "" {
				tmessage = err.Error()
			}
			payReturn.Return.State = customError.IdMessage
			payReturn.Return.Message = customError.Message
			payReturn.Return.TechnicalMessage = tmessage
			return payReturn
		}
		retBuyer = localRetBuyer
	}

	payInfo.Buyer = retBuyer
	//Save general payment info
	payInfo, err := SavePayment(payInfo)
	if err != nil {
		payReturn.Return.State = 0
		payReturn.Return.Message = "Coudln't save payment."
		payReturn.Return.TechnicalMessage = err.Error()
		return payReturn
	}

	//Boleto
	if payInfo.PaymentInfo.PaymentType == 1 {
		//Save at database
		boletoNumber := BoletoPayment()
		payInfo.PaymentInfo.Boleto.Number = boletoNumber
		payReturn.Payment.Boleto.Number = boletoNumber
		payInfo, err = SaveBoletoPayment(payInfo)
		if err != nil {
			payReturn.Return.State = 0
			payReturn.Return.Message = "Coudln't save boleto payment."
			payReturn.Return.TechnicalMessage = err.Error()
			return payReturn
		}

	} else {
		if payInfo.PaymentInfo.PaymentType == 2 {
			//ValidCreditCard
			if ok, msg := ValidCardInfo(payInfo.PaymentInfo.Card); !ok {
				payReturn.Return.State = 0
				payReturn.Return.Message = "Check you card data. " + msg
				payReturn.Return.TechnicalMessage = "Invalid card data."
				return payReturn
			}
			payInfo, err = SaveCardPayment(payInfo)
			if err != nil {
				payReturn.Return.State = 0
				payReturn.Return.Message = "Coudln't save card payment."
				payReturn.Return.TechnicalMessage = err.Error()
				return payReturn
			}

		}
		payReturn.Payment.Card.Successful = CardPayment()
	}
	payReturn.Payment.PaymentId = payInfo.PaymentInfo.PaymentID
	payReturn.Payment.StateId = payInfo.PaymentInfo.PaymentState

	payReturn.Return.State = 1
	payReturn.Return.Message = "Payment processed successfully."
	return payReturn

}

func validatePaymentType(pType int) bool {
	if (pType == 1) || (pType == 2) {
		return true
	} else {
		return false
	}
}

func AlterPaymentState(idPayment int64, idState int) ReturnStruct {
	var ret ReturnStruct
	if ok, msg := AlterPaymentStateDB(idPayment, idState); ok {
		ret.State = 1
		ret.Message = "Payment state altered successfully."
		return ret
	} else {
		ret.State = 0
		ret.Message = "We've couldn't altering payment state."
		ret.TechnicalMessage = msg
		return ret
	}
}

func ConsultPaymentID(idPayment int64) PaymentConsult {
	var retPayment PaymentConsult
	payments, err := ConsultPaymentByIdBD(idPayment)
	if err != nil {
		retPayment.Return.State = 0
		retPayment.Return.Message = "Fail consulting payment."
		retPayment.Return.TechnicalMessage = err.Error()
		return retPayment
		//formatErrorResponse(w, 422, 422, , )
		//return
	}
	if len(payments) < 1 {
		retPayment.Return.State = 0
		retPayment.Return.Message = "Payment not found."
		return retPayment
	}

	retPayment.Return.State = 1
	retPayment.Payments = payments
	return retPayment

}

func ValidCardNumber(cardNumber string) bool {
	return ValidCreditCard(cardNumber)
}

func GetCarBrand(cardNumber string) BrandData {
	brand := GetBrand(cardNumber)
	var brandReturn BrandData
	if brand == Amex {
		brandReturn.Code = Amex
		brandReturn.Name = "Amex"
		return brandReturn
	}
	if brand == Diners {
		brandReturn.Code = Diners
		brandReturn.Name = "Diners"
		return brandReturn
	}
	if brand == Elo {
		brandReturn.Code = Elo
		brandReturn.Name = "Elo"
		return brandReturn
	}
	if brand == Hipercard {
		brandReturn.Code = Hipercard
		brandReturn.Name = "Hipercard"
		return brandReturn
	}
	if brand == Hiper {
		brandReturn.Code = Hiper
		brandReturn.Name = "Hiper"
		return brandReturn
	}
	if brand == Master {
		brandReturn.Code = Master
		brandReturn.Name = "Master"
		return brandReturn
	}
	if brand == Visa {
		brandReturn.Code = Visa
		brandReturn.Name = "Visa"
		return brandReturn
	}

	brandReturn.Code = Unknown
	brandReturn.Name = "Unknown"
	return brandReturn
}
