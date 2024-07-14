package loan_calulator

import (
	"fmt"
	"math"
	"strings"
)

const (
	paymentMessage               = "Your monthly payment = %.2f\n"
	periodsMessage               = "It will take %s to repay the loan\n"
	principalMessage             = "Your loan principal = %.2f!\n"
	differentiatedPaymentMessage = "Month %d: payment is %d\n"
	incorrectParametersMessage   = "Incorrect parameters"
	overpaymentMessage           = "Overpayment = %.2f\n"
)

func Run() {

	var payment, principal, interest, periods, paymentType, err = readInputs()
	if err != nil {
		fmt.Println(err)
		return
	}

	if interest < 0 || paymentType == "" {
		fmt.Println(incorrectParametersMessage)
		return
	}

	option := getCalculationType(paymentType, principal, payment, periods)
	interestRate := interest / (12 * 100)
	var overpayment float64

	switch option {
	case "annuity.periods":
		months := calculateNumberOfPayments(principal, payment, interestRate)
		overpayment = (payment)*float64(months) - principal
		message := formatMonth(months)
		fmt.Printf(periodsMessage, message)
	case "annuity.payment":
		monthPayment := calculateMonthPayment(interestRate, principal, periods)
		overpayment = float64(periods)*(monthPayment) - principal
		fmt.Printf(paymentMessage, monthPayment)
	case "annuity.principal":
		loanPrincipal := calculateLoanPrincipal(interestRate, payment, periods)
		overpayment = (payment)*float64(periods) - loanPrincipal
		fmt.Printf(principalMessage, loanPrincipal)
	case "diff":
		payments := calculateDifferentiatedPayment(interestRate, principal, periods)
		for index, value := range payments {
			overpayment += float64(value)
			fmt.Printf(differentiatedPaymentMessage, index+1, value)
		}
		overpayment = overpayment - principal
	default:
		fmt.Println(incorrectParametersMessage)
		return
	}

	fmt.Printf(overpaymentMessage, overpayment)
}

func calculateMonthPayment(interestRate, principal float64, periods int) float64 {
	payment := principal * (interestRate * math.Pow(1+interestRate, float64(periods)) / (math.Pow(1+interestRate, float64(periods)) - 1))
	return math.Ceil(payment)
}

func calculateLoanPrincipal(interestRate, payment float64, periods int) float64 {
	return payment / (interestRate * math.Pow(1+interestRate, float64(periods)) / (math.Pow(1+interestRate, float64(periods)) - 1))
}

func calculateNumberOfPayments(principal, payment, interestRate float64) int {
	number := math.Log(payment/(payment-interestRate*principal)) / math.Log(1+interestRate)
	return int(math.Ceil(number))
}

func calculateDifferentiatedPayment(interestRate, principal float64, periods int) []int {
	result := make([]int, periods)
	for i := 0; i < periods; i++ {
		p := principal/float64(periods) + interestRate*(principal-principal*float64(i)/float64(periods))
		result[i] = int(math.Ceil(p))
	}
	return result
}

func getCalculationType(paymentType string, principal, payment float64, periods int) string {
	if paymentType == "diff" && principal > 0 && periods > 0 {
		return "diff"
	}
	if principal < 0 && payment > 0 && periods > 0 {
		return "annuity.principal"
	}
	if principal > 0 && payment > 0 && periods < 0 {
		return "annuity.periods"
	}
	if principal > 0 && payment < 0 && periods > 0 {
		return "annuity.payment"
	}
	return ""
}

func formatMonth(months int) string {
	years := months / 12
	reminder := months % 12
	var message string

	switch {
	case years == 0:
		message = fmt.Sprintf("%d months", reminder)
	case reminder == 0:
		message = fmt.Sprintf("%d years", years)
	default:
		message = fmt.Sprintf("%d years and %d months", years, reminder)
	}
	if reminder == 1 {
		message = strings.Replace(message, "months", "month", 1)
	}
	return message
}

func readInputs() (float64, float64, float64, int, string, error) {
	var payment, principal, interest float64
	var periods int
	var paymentType string

	fmt.Print("Enter payment: ")
	_, err := fmt.Scan(&payment)
	if err != nil {
		return 0, 0, 0, 0, "", err
	}

	fmt.Print("Enter principal: ")
	_, err = fmt.Scan(&principal)
	if err != nil {
		return 0, 0, 0, 0, "", err
	}

	fmt.Print("Enter interest: ")
	_, err = fmt.Scan(&interest)
	if err != nil {
		return 0, 0, 0, 0, "", err
	}

	fmt.Print("Enter periods: ")
	_, err = fmt.Scan(&periods)
	if err != nil {
		return 0, 0, 0, 0, "", err
	}

	fmt.Print("Enter payment type: ")
	_, err = fmt.Scan(&paymentType)
	if err != nil {
		return 0, 0, 0, 0, "", err
	}

	return payment, principal, interest, periods, paymentType, nil
}
