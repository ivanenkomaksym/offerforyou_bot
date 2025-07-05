package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot *tgbotapi.BotAPI

	// Повідомлення-привітання, яке бачить користувач на старті
	welcomeMessage = "<b>Привіт!</b> 👋\n\n" +
		"Бачу, ви тут, бо пошук роботи йде не так гладко, як хотілося б. Це нормально, таке буває з кожним.\n\n" +
		"Давайте розберемося, що саме вас турбує. Оберіть ситуацію, яка вам найбільше знайома, і я підкажу, в чому може бути причина та що з цим робити."

	// Тексти відповідей на кожну проблему
	responseRejection = "<b>Проблема:</b> Ви відправляєте десятки резюме, а у відповідь отримуєте або миттєвий реджект (схоже на бота), або просто тишу.\n\n" +
		"<b>Що це означає:</b>\n" +
		"Найімовірніше, ваше резюме не проходить перший фільтр. Воно може бути:\n" +
		"• <b>Нечитабельним</b> для ATS-систем, якими користуються рекрутери.\n" +
		"• <b>Не адаптованим</b> під конкретну вакансію.\n" +
		"• <b>Слабко</b> презентувати ваш досвід і не \"тригерити\" рекрутера.\n\n" +
		"<b>Що робити:</b>\n" +
		"Ваше резюме — це ключ до дверей. Його потрібно професійно оновити та адаптувати, а не сподіватися, що на 20-й раз пощастить."

	responseRecruiterFail = "<b>Проблема:</b> Ви проходите скринінг резюме, доходите до дзвінка з рекрутером, але після нього отримуєте відмову.\n\n" +
		"<b>Що це означає:</b>\n" +
		"Справа не в технічних навичках, а в самопрезентації. Можливі причини:\n" +
		"• Ви не вмієте чітко та структуровано розповісти про свій досвід.\n" +
		"• Звучите невпевнено або, навпаки, зверхньо.\n" +
		"• \"Плаваєте\" в питаннях про зарплатні очікування.\n" +
		"• Не вмієте підтримати розмову і справити гарне перше враження.\n\n" +
		"<b>Що робити:</b>\n" +
		"Проходження інтерв'ю — це навичка, яку можна і треба тренувати. Вас цьому не вчили, тому не соромно звернутися по допомогу."

	responseTechFail = "<b>Проблема:</b> П'ять технічних інтерв'ю — п'ять \"факапів\". Ви відчуваєте, що не дотягуєте саме на цьому етапі.\n\n" +
		"<b>Що це означає:</b>\n" +
		"Проблема не в тому, що ви поганий спеціаліст, а в підході до підготовки та проходження співбесіди. Можливо, ви:\n" +
		"• Губитеся під тиском.\n" +
		"• Неправильно розумієте, що від вас хочуть почути.\n" +
		"• Повторюєте ті самі помилки, не аналізуючи їх.\n\n" +
		"<b>Що робити:</b>\n" +
		"Потрібен системний аналіз. Варто розібрати кожну невдалу співбесіду, знайти слабкі місця та цілеспрямовано їх підтягнути. Іноді для цього потрібен свіжий погляд збоку."

	responseNoResult = "<b>Проблема:</b> Ви оновили резюме, прокачали LinkedIn, пройшли декілька співбесід, але оферу все немає. Час іде, а мотивація тане.\n\n" +
		"<b>Що це означає:</b>\n" +
		"Це найнебезпечніший дзвіночок. Ви потрапили в замкнене коло, і самостійні дії не дають результату. Це призводить до вигорання та зневіри у власних силах.\n\n" +
		"<b>Що робити:</b>\n" +
		"Вам потрібна не просто порада, а <b>комплексний супровід</b>. Потрібен професіонал, який зануриться у вашу ситуацію, знайде корінь проблеми, розробить стратегію і буде поруч на кожному кроці до отримання оферу."

	// --- Тексти кнопок та дані для Callback ---

	// Головне меню
	btnTextRejection = "Відгук ➞ Відмова"
	btnDataRejection = "rejection_fail" // callback_data

	btnTextRecruiterFail = "Провалюю співбесіду з рекрутером"
	btnDataRecruiterFail = "recruiter_fail" // callback_data

	btnTextTechFail = "Провалюю технічну співбесіду"
	btnDataTechFail = "tech_fail" // callback_data

	btnTextNoResult = "Роблю все, а результату немає"
	btnDataNoResult = "no_result" // callback_data

	// Навігація та контакти
	btnTextBack = "⬅️ Назад до меню"
	btnDataBack = "back_to_main" // callback_data

	btnTextContact = "✍️ Зв'язатися з консультантом"
	contactURL     = "https://t.me/Anastasiia_hrg" // Пряме посилання на контакт

	// --- Клавіатури (Markup) ---

	// Клавіатура для головного меню. 4 кнопки, розташовані у 2 ряди
	mainMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnTextRejection, btnDataRejection),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnTextRecruiterFail, btnDataRecruiterFail),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnTextTechFail, btnDataTechFail),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnTextNoResult, btnDataNoResult),
		),
	)

	// Клавіатура для меню відповіді. Дві кнопки: одна для контакту, інша для повернення назад
	responseMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(btnTextContact, contactURL),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnTextBack, btnDataBack),
		),
	)
)

func main() {
	var err error

	godotenv.Load()

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN environment variable not set.")
	}

	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// Set debug mode if DEBUG environment variable is "true"
	debugEnv := os.Getenv("DEBUG")
	bot.Debug = strings.ToLower(debugEnv) == "true"

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Get the WEBHOOK_URL from environment variables
	// This will be the URL of your deployed Cloud Run service + the webhook path
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("WEBHOOK_URL environment variable not set. This should be your Cloud Run service URL including the path (e.g., https://<service-url>/telegram-webhook).")
	}

	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		log.Fatalf("Failed to create webhook config: %v", err)
	}
	// Use bot.Request to send the WebhookConfig to Telegram
	_, err = bot.Request(wh)
	if err != nil {
		log.Fatalf("Failed to set webhook: %v", err)
	}

	// Get webhook info to confirm it's set and check for any errors from Telegram's side
	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatalf("Failed to get webhook info: %v", err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram last webhook error: %s", info.LastErrorMessage)
	}
	log.Printf("Webhook set to: %s (pending: %t)", info.URL, info.PendingUpdateCount > 0)

	// Get the port from environment variables, default to 8080 for Cloud Run
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// Create a new HTTP multiplexer
	mux := http.NewServeMux()

	// Define the webhook endpoint that Telegram will send updates to
	// It's good practice to use a non-root path for webhooks.
	mux.HandleFunc("/telegram-webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
			return
		}

		var update tgbotapi.Update
		// Decode the JSON request body into a tgbotapi.Update struct
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			log.Printf("ERROR: Could not decode update: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Process the update in a non-blocking way if possible,
		// but for simple bots, direct handling is fine.
		handleUpdate(update)

		// Respond with 200 OK to Telegram immediately
		// This acknowledges receipt of the update and prevents Telegram from retrying.
		w.WriteHeader(http.StatusOK)
	})

	// Add a health check endpoint for Cloud Run
	// Cloud Run sends requests to the root path by default for health checks.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Starting HTTP server on %s", addr)
	// Start the HTTP server. This will block indefinitely, serving requests.
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func handleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		user := update.CallbackQuery.From
		text := update.CallbackQuery.Data

		log.Printf("'[%s] %s %s' selected '%s'", user.UserName, user.FirstName, user.LastName, text)

		// Створюємо повідомлення для редагування
		// bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
		msg := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			"", // Текст буде встановлено нижче
		)

		// Визначаємо, яка клавіатура буде у відповіді
		var markup tgbotapi.InlineKeyboardMarkup

		switch update.CallbackQuery.Data {
		case btnDataRejection:
			msg.Text = responseRejection
			markup = responseMarkup
		case btnDataRecruiterFail:
			msg.Text = responseRecruiterFail
			markup = responseMarkup
		case btnDataTechFail:
			msg.Text = responseTechFail
			markup = responseMarkup
		case btnDataNoResult:
			msg.Text = responseNoResult
			markup = responseMarkup
		case btnDataBack:
			msg.Text = welcomeMessage
			markup = mainMenuMarkup
		}

		msg.ReplyMarkup = &markup
		msg.ParseMode = "HTML" // Важливо, щоб теги <b> працювали
		bot.Send(msg)
	}

	// Якщо це команда /start
	if update.Message != nil && update.Message.IsCommand() {
		user := update.Message.From
		log.Printf("'[%s] %s %s' started chat", user.UserName, user.FirstName, user.LastName)

		if update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMessage)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = mainMenuMarkup
			bot.Send(msg)
		}
	}
}
