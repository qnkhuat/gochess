Knock out all of your opponents to win!

# Garbage

To attack opponents, lines with random holes are sent to fill their playfields.

# Target

Garbage is sent to the opponent who has received the least amount garbage from
anyone.

# Combo

Clearing one or more lines will increment a counter by one and add time to a
countdown. The counter is reset to zero when the countdown expires.

Two time values are calculated and added to the countdown, **base time** and
**bonus time**. Base time is 2.4 seconds, and bonus time is half of that. Each
time value is halved as the counter increases. Bonus time is multiplied by the
number of lines being cleared. **It is more effective to start a combo with a
multi-line clear.**

When at least two lines are cleared, (cleared - 1) lines of garbage are sent.
Bonus garbage is sent each time the counter increases.

| Counter | Base time | Bonus time | Bonus garbage |
|---|---|---|---|
1 | 2.4 | 1.2 | 0
2 | 1.2 | 0.6 | 1
3 | 0.6 | 0.3 | 1
4 | 0.3 | 0.15 | 1
5 | 0.15 | 0.075 | 1
6 | 0.075 | 0.038 | 2
7 | 0.038 | 0.019 | 3
8 | 0.019 | 0.009 | 5
9 | 0.009 | 0.005 | 8
10 | 0.005 | 0.002 | 13
