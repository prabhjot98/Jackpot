# Jackpot

This is a (fake) slot machine made to be played on terminals. I made this game because I wanted to do something fun
at the end of a work day. Currently, it's in beta because the gameplay loop needs some more polish and love, but I'd like
to release the 1.0 version sometime this year.

## Install

```
sudo curl -L https://raw.githubusercontent.com/prabhjot98/jackpot/refs/heads/main/install.sh | bash
```

## How to play

Press the space bar to spin the slot machine and 'q' to quit the game. The game saves itself when you quit so you won't
have to worry about losing your progress.

### Mechanics of the game

1. You spend one token to spin the slot machine.
2. You get awarded 40 play tokens everyday (resets 12:00am in your local time).
3. Everytime you lose you get $5 added to the **jackpot**.

### Symbols and what they do

3️⃣  You win 3 dollars

5️⃣  You win 5 dollars

7️⃣  You win 7 dollars

🔟 You win 10 dollars

💯 You win 100 dollars

🆓 You get a free spin

🍯 You win all the money in the jackpot

🃏 It matches any 3, 5, or 7 on the slot

🎲 You get to reroll your multiplier die

🌕 Day turns to night

🌅 Night turns back to day

🔥 You enter fever mode

💀 You get to spin on the Wheel of Misfortune

### Glossary

Fever mode: Your next win on the machine is doubled.

Multiplier: You have a base multiplier of x1. Whatever you win is multiplied by this multiplier except for the jackpot and 💯. The multiplier can change if you get 3 dice. You then get to roll a die and the result becomes your new multiplier

Night time: 4x the probability of rolling a joker and 2x the probability of rolling a skull

Day time: The normal mode for the game

Jackpot: Everytime you lose on the machine you get $5 added to the slot machine

Wheel of Misfortune: A wheel where you can get bad results. The bad results can range from losing a single dollar to losing everything!
