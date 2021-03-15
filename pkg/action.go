package pkg

type Action string

const (
	ActionDrawOffer     Action = "Want Draw"
	ActionDrawPrompt           = "Draw?"
	ActionDrawAccept           = "Accept"
	ActionDrawReject           = "Reject"
	ActionResignPrompt         = "Resign"
	ActionResignYes            = "Yes"
	ActionResignNo             = "No"
	ActionNewGamePrompt        = "New Game?"
	ActionNewGameOffer         = "New Game"
	ActionNewGameAccept        = "Yes!"
	ActionNewGameReject        = "No~"
	ActionExit                 = "Exit"
	ActionWin                  = "Win"
	ActionLose                 = "Lose"
	ActionDraw                 = "Draw"
)
