package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"github.com/joho/godotenv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (

	// Store bot screaming status
	screaming = false
	bot       *tgbotapi.BotAPI

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
	btnTextRejection      = "Відгук ➞ Відмова"
	btnDataRejection      = "rejection_fail" // callback_data
	
	btnTextRecruiterFail  = "Провалюю співбесіду з рекрутером"
	btnDataRecruiterFail  = "recruiter_fail" // callback_data
	
	btnTextTechFail       = "Провалюю технічну співбесіду"
	btnDataTechFail       = "tech_fail"      // callback_data
	
	btnTextNoResult       = "Роблю все, а результату немає"
	btnDataNoResult       = "no_result"      // callback_data
	
	// Навігація та контакти
	btnTextBack           = "⬅️ Назад до меню"
	btnDataBack           = "back_to_main"   // callback_data
	
	btnTextContact        = "✍️ Зв'язатися з консультантом"
	contactURL            = "https://t.me/Anastasiia_hrg" // Пряме посилання на контакт
	
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

	err = godotenv.Load()
    if err != nil {
    	log.Fatal(err)
    }

	bot_token := os.Getenv("BOT_TOKEN")
	bot, err = tgbotapi.NewBotAPI(bot_token)
	if err != nil {
		// Abort if something is wrong
		log.Panic(err)
	}

	// Set this to true to log all interactions with telegram servers
	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Create a new cancellable background context. Calling `cancel()` leads to the cancellation of the context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)

	// Pass cancellable context to goroutine
	go receiveUpdates(ctx, updates)

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press enter to stop")

	// Wait for a newline symbol, then cancel handling updates
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()

}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			handleUpdate(update)
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
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
		if update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMessage)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = mainMenuMarkup
			bot.Send(msg)
		}
	}
}

func handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		err = handleCommand(message.Chat.ID, text)
	} else if screaming && len(text) > 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, strings.ToUpper(text))
		// To preserve markdown, we attach entities (bold, italic..)
		msg.Entities = message.Entities
		_, err = bot.Send(msg)
	} else {
		// This is equivalent to forwarding, without the sender's name
		copyMsg := tgbotapi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
		_, err = bot.CopyMessage(copyMsg)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

// When we get a command, we react accordingly
func handleCommand(chatId int64, command string) error {
	var err error

	switch command {
	case "/scream":
		screaming = true
		break

	case "/whisper":
		screaming = false
		break
	}

	return err
}