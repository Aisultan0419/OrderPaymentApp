package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/payments": {
            "post": {
                "description": "Processes a payment for an order. Amounts \u003e 100000 cents ($1000) are automatically Declined.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "payments"
                ],
                "summary": "Authorize a payment",
                "parameters": [
                    {
                        "description": "Payment payload",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/http.authorizePaymentRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/http.paymentResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/http.errorResponse"
                        }
                    }
                }
            }
        },
        "/payments/{order_id}": {
            "get": {
                "description": "Returns the payment record associated with the given order ID.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "payments"
                ],
                "summary": "Get payment by order ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Order ID",
                        "name": "order_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/http.paymentResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/http.errorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "http.authorizePaymentRequest": {
            "type": "object",
            "required": [
                "amount",
                "order_id"
            ],
            "properties": {
                "amount": {
                    "type": "integer"
                },
                "order_id": {
                    "type": "string"
                }
            }
        },
        "http.errorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "http.paymentResponse": {
            "type": "object",
            "properties": {
                "amount": {
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "order_id": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                },
                "transaction_id": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8081",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Payment Service API",
	Description:      "Processes payments and validates transaction limits. Amounts above $1000 (100000 cents) are declined.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
