A matching engine matches orders from market participants. These matches will result in trades. A trade occurs when Participant A
wants to buy a particular asset at an equal or higher price than Participant B is willing to sell that same asset.
When someone wants to buy an asset, a market participant sends a buy order for a given symbol (e.g. FFLY). A sent order contains an id, symbol, side, limit price and volume. The limit price indicates that in the case of a buy order, you are willing to buy at the given price or lower. In the case of a sell order, the limit price indicates that you are willing to sell at the given price or higher.

All orders are managed in a central limit order book which has two sides, the buy side and the sell side. If a new order is received by the matching engine, it first checks whether it can match with orders already in the order book on the opposite side.
The order will be matched with the opposite side until either the volume of the new order is exhausted or until there are no orders on the opposite side with which the new order can match. The matching priority starts with price priority; the "best" price matches first. If there are multiple orders in the order book at the same price, these orders are matched with time priority; orders
were inserted earlier are matched first.

Two operations can be applied to an order once it is in the order book; "cancel' and "update." A cancel removes the order from the order book. An update changes the price and/or volume of the order. An update causes the order to lose time priority in
the order book, unless the only change to the order is that the volume is decreased. If the order price is updated, it needs to be re-evaluated for potential matches.


Examples
Suppose the order book has the following open orders:

## Orderbook
ID Side Volume Price
1 Buy 1 122
2 Buy
4 Buy
23
3 Buy 7
120
121 ‹- Higher time priority than the order below since it came in earlier
121 ‹- Higher price priority than the order below since it offers a better price
Next, we insert a sell order (ID=5) with a volume 16 and price 120. The order book now looks like this:

## Fills
1
1
2
12
4
3
MatchedId Volume Price (in the following order)
122
121
121
## Orderbook
ID Side Volume Price
4 Buy 20
121
3 Buy 7
120
Finally we insert another sell order (ID=6) with volume 24 and price 121. The order book now looks like this:
## Fills
MatchedId Volume Price
4
20
121
## Orderbook
ID Side Volume Price
6 Sell 4
3 Buy 7
120
121 <- leftover volume after match


## The challenging part is this to be honest

 The expected output is:
 - List of trades in chronological order with the format:
   <symbol>,<price>,<volume>,<taker_order_id>,<maker_order_id>
   e.g. FFLY,23.55,11,4,7
   The maker order is the one being removed from the order book, the taker order is the incoming one matching it.
 - Then, per symbol (in alphabetical order):
   - separator "===<symbol>==="
   - bid and ask price levels (sorted best to worst by price) for that symbol in the format:
     SELL,<ask_price>,<ask_volume>
     SELL,<ask_price>,<ask_volume>
     BUY,<bid_price>,<bid_volume>
     BUY,<bid_price>,<bid_volume>
     e.g. SELL,25.67,102
          SELL,25.56,34
          BUY,25.52,23
          BUY,25.51,11
          BUY,25.43,4


We are able to succesfully match orders and make trades on one symbol. That is a good achievement so far, but what if we want to expand that to include 1000 symbol, how can we go about that:
- create a map of order books, where each symbol has its own order book -- trades
- another important consideration is to sync the access of entries to whatever group we want
- one approach, is maybe like this:
    - on passing from sell / buy -- to update and cancel, we send the paricular order to map[[]trades]
    - order_id is provided, so we won't have to worry about ordering
    - now, each item belongs to a dictionry
    - now, do i need to have a lock and a global order book to store all of the trades, or not? it feels yes, but i am not sure



Expected [
    FFLY,46,3,2,4
    FFLY,45.95,1,5,1
    FFLY,45.95,1,6,1
    FFLY,45.95,1,7,3
    ===FFLY===
    SELL,46,5 
    BUY,45.95,16

    ], but got
    
    FFLY,45.95,1,5,1 
    FFLY,45.95,1,6,1 
    FFLY,45.95,1,7,1 
    ===FFLY=== 
    SELL,46.00,8 
    BUY,45.95,2 BUY,45.95,6 
    BUY,45.95,12
    ]


[FFLY,46,3,2,4 FFLY,45.95,1,5,1 FFLY,45.95,1,6,1 FFLY,45.95,1,7,3 ===FFLY=== SELL,46,5 BUY,45.95,16]

, but got

[FFLY,46,3,2,4 FFLY,45.95,1,5,1 FFLY,45.95,1,6,1 FFLY,45.95,1,7,1 ===FFLY=== SELL,46,5 BUY,45.95,4 BUY,45.95,12]
FAIL

[FFLY,46,3,2,4 FFLY,45.95,1,5,1 FFLY,45.95,1,6,1 FFLY,45.95,1,7,3 ===FFLY=== SELL,46,5 BUY,45.95,16]

[FFLY,46,3,2,4 ===FFLY=== SELL,46,5 SELL,45.95,1 SELL,45.95,1 SELL,45.95,1 BUY,45.95,5 BUY,45.95,12]

Running tool: /usr/local/go/bin/go test -timeout 30s -run ^TestRunMatchingEngine$ github.com/adonese/bluefin

[FFLY,46,3,2,4 FFLY,45.95,1,5,1 FFLY,45.95,1,6,1 FFLY,45.95,1,7,3 ===FFLY=== SELL,46,5 BUY,45.95,16], but got 
[FFLY,46,3,2,4 ===FFLY=== SELL,46,5 SELL,45.95,1 SELL,45.95,1 SELL,45.95,1 BUY,45.95,5 BUY,45.95,12]



--- FAIL: TestRunMatchingEngine (0.00s)
    --- FAIL: TestRunMatchingEngine/Test_Case_5 (0.00s)
        /Users/adonese/src/seed/main_test.go:102: Expected 
[FFLY,46,3,2,4 FFLY,45.95,1,5,1 FFLY,45.95,1,6,1 FFLY,45.95,1,7,3 ===FFLY=== SELL,46,5 BUY,45.95,16]
, but got 
[FFLY,46,3,2,4 ===FFLY=== SELL,46,5 SELL,45.95,1 SELL,45.95,1 SELL,45.95,1 BUY,45.95,5 BUY,45.95,12]
    --- FAIL: TestRunMatchingEngine/Test_Case_4 (0.00s)
        /Users/adonese/src/seed/main_test.go:102: Expected [FFLY,14.235,6,8,2 FFLY,14.235,12,8,3 FFLY,14.234,5,8,4 ===FFLY=== SELL,14.24,9 SELL,14.237,8 SELL,14.234,2 BUY,14.23,3], but got [===FFLY=== SELL,14.24,9 SELL,14.237,8 SELL,14.234,25 BUY,14.235,6 BUY,14.234,5 BUY,14.235,12 BUY,14.23,3]



Expected 
[FFLY,14.235,6,8,2 FFLY,14.235,12,8,3 FFLY,14.234,5,8,4 ===FFLY=== SELL,14.24,9 SELL,14.237,8 SELL,14.234,2 BUY,14.23,3], but got 
[FFLY,14.235,6,8,2 FFLY,14.235,12,8,3 FFLY,14.234,5,8,4 ===FFLY=== SELL,14.234,2 SELL,14.24,9 SELL,14.237,8 BUY,14.23,3]

[FFLY,14.235,6,8,2 FFLY,14.235,12,8,3 FFLY,14.234,5,8,4 ===FFLY=== SELL,14.24,9 SELL,14.237,8 SELL,14.234,2 BUY,14.23,3], but got 
[FFLY,14.235,6,8,2 FFLY,14.235,12,8,3 FFLY,14.234,5,8,4 ===FFLY=== SELL,14.237,8 SELL,14.24,9 SELL,14.234,2 BUY,14.23,3]

Expected 
[FFLY,14.235,6,8,2 FFLY,14.235,12,8,3 FFLY,14.234,5,8,4 ===FFLY=== SELL,14.24,9 SELL,14.237,8 SELL,14.234,2 BUY,14.23,3], but got 
[FFLY,14.235,6,8,2 FFLY,14.235,12,8,3 FFLY,14.234,5,8,4 ===FFLY=== SELL,14.234,2 SELL,14.237,8 SELL,14.24,9 BUY,14.23,3]



FAILING TEST 6

Operations AREEEEEEEEE======: 

[INSERT,1,FFLY,SELL,12.2,5 INSERT,2,FFLY,SELL,12.1,8, INSERT,3,FFLY,BUY,12.5,10]


FAIling test 9
Operations AREEEEEEEEE======: 
[INSERT,1,FFLY,BUY,47,5 INSERT,2,FFLY,BUY,47,6 INSERT,3,FFLY,SELL,47,9 UPDATE,2,47,-1]


FAILING TEST 11
Operations AREEEEEEEEE======: 
[INSERT,1,FFLY,BUY,47,5 INSERT,2,FFLY,BUY,47,6 INSERT,3,FFLY,SELL,47,9 UPDATE,1,45,2 UPDATE,5,45,2]

