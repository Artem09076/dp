package templates

import (
	"bytes"
	"html/template"
)

type EmailData struct {
	Title        string
	Username     string
	ButtonText   string
	ButtonURL    string
	Year         int
	SupportEmail string
}

type BookingData struct {
	EmailData
	ServiceName string
	DateTime    string
}

type ProfileData struct {
	EmailData
	VerificationStatus string
	Reason             string
}

func RenderBookingCreated(data BookingData) (string, error) {
	htmlContent := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f5;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            padding: 0;
            background-color: #ffffff;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: #667eea;
            color: white !important;
            text-decoration: none;
            border-radius: 6px;
            margin: 20px 0;
            font-weight: 500;
        }
        .button:hover {
            background-color: #5a67d8;
        }
        .info-box {
            background-color: #f7fafc;
            border-left: 4px solid #667eea;
            padding: 15px 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .info-item {
            margin: 10px 0;
        }
        .info-label {
            font-weight: 600;
            color: #4a5568;
        }
        .footer {
            background-color: #f7fafc;
            padding: 20px 30px;
            text-align: center;
            font-size: 12px;
            color: #718096;
            border-top: 1px solid #e2e8f0;
        }
        @media (max-width: 600px) {
            .content {
                padding: 25px 20px;
            }
            .header h1 {
                font-size: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        <div class="content">
            {{if .Username}}<p>Здравствуйте, <strong>{{.Username}}</strong>!</p>{{end}}
            
            <p>Ваша услуга заинтересовала нового клиента!</p>
            
            <div class="info-box">
                <div class="info-item">
                    <span class="info-label">Услуга:</span> {{.ServiceName}}
                </div>
                <div class="info-item">
                    <span class="info-label">Дата и время:</span> {{.DateTime}}
                </div>
            </div>
            
            <p>Чтобы подтвердить бронирование, пожалуйста, нажмите на кнопку ниже.</p>
            <p><strong>Важно:</strong> Если вы не подтвердите бронь в течение 24 часов, она будет автоматически отменена.</p>
            
            {{if .ButtonText}}
            <div style="text-align: center;">
                <a href="{{.ButtonURL}}" class="button">{{.ButtonText}}</a>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>Это автоматическое сообщение, пожалуйста, не отвечайте на него.</p>
            <p>© {{.Year}} Ваш Сервис. Все права защищены.</p>
            <p>По вопросам обращайтесь: <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("booking_created").Parse(htmlContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderBookingSubmit(data BookingData) (string, error) {
	htmlContent := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f5;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            padding: 0;
            background-color: #ffffff;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: #48bb78;
            color: white !important;
            text-decoration: none;
            border-radius: 6px;
            margin: 20px 0;
            font-weight: 500;
        }
        .info-box {
            background-color: #f7fafc;
            border-left: 4px solid #48bb78;
            padding: 15px 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .info-item {
            margin: 10px 0;
        }
        .info-label {
            font-weight: 600;
            color: #4a5568;
        }
        .footer {
            background-color: #f7fafc;
            padding: 20px 30px;
            text-align: center;
            font-size: 12px;
            color: #718096;
            border-top: 1px solid #e2e8f0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        <div class="content">
            {{if .Username}}<p>Здравствуйте, <strong>{{.Username}}</strong>!</p>{{end}}
            
            <p>Ваше бронирование успешно подтверждено!</p>
            
            <div class="info-box">
                <div class="info-item">
                    <span class="info-label">Услуга:</span> {{.ServiceName}}
                </div>
                <div class="info-item">
                    <span class="info-label">Дата и время:</span> {{.DateTime}}
                </div>
            </div>
            
            <p>Мы желаем вам приятно провести время! Если у вас возникнут вопросы, свяжитесь с нами.</p>
            
            {{if .ButtonText}}
            <div style="text-align: center;">
                <a href="{{.ButtonURL}}" class="button">{{.ButtonText}}</a>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>Это автоматическое сообщение, пожалуйста, не отвечайте на него.</p>
            <p>© {{.Year}} Ваш Сервис. Все права защищены.</p>
            <p>По вопросам обращайтесь: <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("booking_submit").Parse(htmlContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderBookingCancelled(data BookingData) (string, error) {
	htmlContent := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f5;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            padding: 0;
            background-color: #ffffff;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #f56565 0%, #e53e3e 100%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: #667eea;
            color: white !important;
            text-decoration: none;
            border-radius: 6px;
            margin: 20px 0;
            font-weight: 500;
        }
        .info-box {
            background-color: #fff5f5;
            border-left: 4px solid #f56565;
            padding: 15px 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .info-item {
            margin: 10px 0;
        }
        .info-label {
            font-weight: 600;
            color: #4a5568;
        }
        .footer {
            background-color: #f7fafc;
            padding: 20px 30px;
            text-align: center;
            font-size: 12px;
            color: #718096;
            border-top: 1px solid #e2e8f0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        <div class="content">
            {{if .Username}}<p>Здравствуйте, <strong>{{.Username}}</strong>!</p>{{end}}
            
            <p>Ваше бронирование было <strong style="color: #e53e3e;">отменено</strong>.</p>
            
            <div class="info-box">
                <div class="info-item">
                    <span class="info-label">Услуга:</span> {{.ServiceName}}
                </div>
                <div class="info-item">
                    <span class="info-label">Дата и время:</span> {{.DateTime}}
                </div>
            </div>
            
            <p>Если вы не инициировали отмену или у вас есть вопросы, пожалуйста, свяжитесь с нашей службой поддержки.</p>
            
            {{if .ButtonText}}
            <div style="text-align: center;">
                <a href="{{.ButtonURL}}" class="button">{{.ButtonText}}</a>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>Это автоматическое сообщение, пожалуйста, не отвечайте на него.</p>
            <p>© {{.Year}} Ваш Сервис. Все права защищены.</p>
            <p>По вопросам обращайтесь: <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("booking_cancelled").Parse(htmlContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderBookingUpdated(data BookingData, version int) (string, error) {
	var content string

	if version == 1 {
		content = `
            <p>Время вашей услуги было изменено.</p>
            
            <div class="info-box">
                <div class="info-item">
                    <span class="info-label">Услуга:</span> {{.ServiceName}}
                </div>
                <div class="info-item">
                    <span class="info-label">Новое время:</span> <strong>{{.DateTime}}</strong>
                </div>
            </div>
            
            <p>Если новое время вам не подходит, вы можете связаться с исполнителем или отменить бронирование.</p>`
	} else {
		content = `
            <p>Исполнитель предложил новое время для вашей услуги.</p>
            
            <div class="info-box">
                <div class="info-item">
                    <span class="info-label">Услуга:</span> {{.ServiceName}}
                </div>
                <div class="info-item">
                    <span class="info-label">Предлагаемое время:</span> <strong>{{.DateTime}}</strong>
                </div>
            </div>
            
            <p>Пожалуйста, подтвердите новое время, если оно вас устраивает. В противном случае бронирование будет автоматически отменено.</p>`
	}

	htmlContent := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f5;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            padding: 0;
            background-color: #ffffff;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #ed8936 0%, #dd6b20 100%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: #ed8936;
            color: white !important;
            text-decoration: none;
            border-radius: 6px;
            margin: 20px 0;
            font-weight: 500;
        }
        .info-box {
            background-color: #fffaf0;
            border-left: 4px solid #ed8936;
            padding: 15px 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .info-item {
            margin: 10px 0;
        }
        .info-label {
            font-weight: 600;
            color: #4a5568;
        }
        .footer {
            background-color: #f7fafc;
            padding: 20px 30px;
            text-align: center;
            font-size: 12px;
            color: #718096;
            border-top: 1px solid #e2e8f0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        <div class="content">
            {{if .Username}}<p>Здравствуйте, <strong>{{.Username}}</strong>!</p>{{end}}
            ` + content + `
            {{if .ButtonText}}
            <div style="text-align: center;">
                <a href="{{.ButtonURL}}" class="button">{{.ButtonText}}</a>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>Это автоматическое сообщение, пожалуйста, не отвечайте на него.</p>
            <p>© {{.Year}} Ваш Сервис. Все права защищены.</p>
            <p>По вопросам обращайтесь: <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("booking_updated").Parse(htmlContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// templates/email_templates.go (добавьте эти функции в существующий файл)

// RenderProfileVerificationSuccess рендерит письмо об успешной верификации
func RenderProfileVerificationSuccess(data ProfileData) (string, error) {
	htmlContent := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f5;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            padding: 0;
            background-color: #ffffff;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: #48bb78;
            color: white !important;
            text-decoration: none;
            border-radius: 6px;
            margin: 20px 0;
            font-weight: 500;
        }
        .button:hover {
            background-color: #38a169;
        }
        .success-box {
            background-color: #f0fff4;
            border-left: 4px solid #48bb78;
            padding: 20px;
            margin: 20px 0;
            border-radius: 8px;
        }
        .success-icon {
            font-size: 48px;
            text-align: center;
            margin-bottom: 10px;
        }
        .feature-list {
            list-style: none;
            padding: 0;
            margin: 15px 0;
        }
        .feature-list li {
            padding: 8px 0;
            padding-left: 25px;
            position: relative;
        }
        .feature-list li:before {
            content: "✓";
            color: #48bb78;
            font-weight: bold;
            position: absolute;
            left: 0;
        }
        .footer {
            background-color: #f7fafc;
            padding: 20px 30px;
            text-align: center;
            font-size: 12px;
            color: #718096;
            border-top: 1px solid #e2e8f0;
        }
        @media (max-width: 600px) {
            .content {
                padding: 25px 20px;
            }
            .header h1 {
                font-size: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        <div class="content">
            {{if .Username}}<p>Здравствуйте, <strong>{{.Username}}</strong>!</p>{{end}}
            
            <div class="success-icon">🎉</div>
            
            <div class="success-box">
                <p style="font-size: 18px; font-weight: 600; color: #22543d; text-align: center; margin: 0 0 10px 0;">
                    Ваш профиль успешно верифицирован!
                </p>
                <p style="text-align: center;">Теперь вам доступны все функции сервиса</p>
            </div>
            
            <p>Спасибо, что выбрали наш сервис! Мы уверены, что сотрудничество с нами будет приятным и продуктивным.</p>
            
            {{if .ButtonText}}
            <div style="text-align: center;">
                <a href="{{.ButtonURL}}" class="button">{{.ButtonText}}</a>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>Это автоматическое сообщение, пожалуйста, не отвечайте на него.</p>
            <p>© {{.Year}} Ваш Сервис. Все права защищены.</p>
            <p>По вопросам обращайтесь: <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("profile_success").Parse(htmlContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderProfileVerificationReject рендерит письмо об отклонении верификации
func RenderProfileVerificationReject(data ProfileData) (string, error) {
	htmlContent := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f5;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            padding: 0;
            background-color: #ffffff;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #f56565 0%, #e53e3e 100%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: #667eea;
            color: white !important;
            text-decoration: none;
            border-radius: 6px;
            margin: 20px 0;
            font-weight: 500;
        }
        .button:hover {
            background-color: #5a67d8;
        }
        .error-box {
            background-color: #fff5f5;
            border-left: 4px solid #f56565;
            padding: 20px;
            margin: 20px 0;
            border-radius: 8px;
        }
        .error-icon {
            font-size: 48px;
            text-align: center;
            margin-bottom: 10px;
        }
        .reason-box {
            background-color: #fed7d7;
            padding: 15px;
            border-radius: 6px;
            margin: 15px 0;
        }
        .action-buttons {
            display: flex;
            gap: 15px;
            justify-content: center;
            margin: 25px 0;
            flex-wrap: wrap;
        }
        .button-secondary {
            display: inline-block;
            padding: 12px 30px;
            background-color: #718096;
            color: white !important;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
        }
        .footer {
            background-color: #f7fafc;
            padding: 20px 30px;
            text-align: center;
            font-size: 12px;
            color: #718096;
            border-top: 1px solid #e2e8f0;
        }
        @media (max-width: 600px) {
            .content {
                padding: 25px 20px;
            }
            .header h1 {
                font-size: 20px;
            }
            .action-buttons {
                flex-direction: column;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        <div class="content">
            {{if .Username}}<p>Здравствуйте, <strong>{{.Username}}</strong>!</p>{{end}}
            
            <div class="error-icon">😔</div>
            
            <div class="error-box">
                <p style="font-size: 18px; font-weight: 600; color: #742a2a; text-align: center; margin: 0 0 10px 0;">
                    К сожалению, ваш профиль не прошел верификацию
                </p>
                
                {{if .Reason}}
                <div class="reason-box">
                    <strong>Причина отклонения:</strong><br>
                    {{.Reason}}
                </div>
                {{end}}
                
                <p><strong>Что делать?</strong></p>
                <ul>
                    <li>Проверьте правильность заполненных данных</li>
                </ul>
            </div>
            
            <p>Если вы не согласны с решением, свяжитесь со службой поддержки</p>
            
            <div class="action-buttons">
                {{if .ButtonText}}
                <a href="{{.ButtonURL}}" class="button">{{.ButtonText}}</a>
                {{end}}
                <a href="mailto:{{.SupportEmail}}" class="button-secondary">Связаться с поддержкой</a>
            </div>
        </div>
        <div class="footer">
            <p>Это автоматическое сообщение, пожалуйста, не отвечайте на него.</p>
            <p>© {{.Year}} Сервис. Все права защищены.</p>
            <p>По вопросам обращайтесь: <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("profile_reject").Parse(htmlContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
