#!/usr/bin/env bash

denom="denom_basic"
user1=$(mpcli keys show user1 -a)
user2=$(mpcli keys show user2 -a)
sellerBeneficiary=$(mpcli keys show sellerBeneficiary -a)
buyerBeneficiary=$(mpcli keys show buyerBeneficiary -a)

echo "User 1: $user1"
echo "User 2: $user2"
echo "====================================================================="

# Generate 5 tokens.
for i in 1 2 3 4 5 6 7
do
	echo "Minting token $i..."
	mpcli tx nft mint ${denom} "TOKEN_$i" ${user1} --from user1 -y <<< '12345678'
	sleep 5
done

# >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

echo "Transfer TOKEN_1 to user2"
mpcli tx nft transfer ${user1} ${user2} ${denom} TOKEN_1 --from user1 -y <<< '12345678'
sleep 5

# >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

echo "Put TOKEN_2 on market"
mpcli tx marketplace put_on_market TOKEN_2 100token ${sellerBeneficiary} --from user1 -y <<< '12345678'
sleep 5

echo "Remove TOKEN_2 from market"
mpcli tx marketplace remove_from_market TOKEN_2 --from user1 -y <<< '12345678'
sleep 5

# >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

echo  "Put TOKEN_2 back on market"
mpcli tx marketplace put_on_market TOKEN_2 100token ${sellerBeneficiary} --from user1 -y <<< '12345678'
sleep 5

echo "Buy TOKEN_2"
mpcli tx marketplace buy TOKEN_2 ${buyerBeneficiary} --from user2 -y <<< '12345678'
sleep 5

# >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

echo "Put TOKEN_3 on auction"
mpcli tx marketplace put_on_auction TOKEN_3 10token ${sellerBeneficiary} 10m --from user1 -y <<< '12345678'
sleep 5

echo "Remove TOKEN_3 from auction"
mpcli tx marketplace remove_from_auction TOKEN_3 --from user1 -y <<< '12345678'
sleep 5

# >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

echo "Put TOKEN_3 on auction again"
mpcli tx marketplace put_on_auction TOKEN_3 10token ${sellerBeneficiary} 10m --buyout 100token --from user1 -y <<< '12345678'
sleep 5

echo "Bid on TOKEN_3"
mpcli tx marketplace bid TOKEN_3 ${buyerBeneficiary} 50token --from user2 -y <<< '12345678'
sleep 5

echo "Bid on TOKEN_3 a sum greater than buyout (results in token ownership change)"
mpcli tx marketplace bid TOKEN_3 ${buyerBeneficiary} 110token --from user2 -y <<< '12345678'
sleep 5

# >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

echo "Put TOKEN_4 on auction"
mpcli tx marketplace put_on_auction TOKEN_4 10token ${sellerBeneficiary} 10m --buyout 100token --from user1 -y <<< '12345678'
sleep 5

echo "Buyout TOKEN_4"
mpcli tx marketplace buyout TOKEN_4 ${sellerBeneficiary} --from user2 -y <<< '12345678'
sleep 5

# >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

echo "Put TOKEN_5 on auction"
mpcli tx marketplace put_on_auction TOKEN_5 10token ${sellerBeneficiary} 10m --buyout 100token --from user1 -y <<< '12345678'
sleep 5

echo "Bid on TOKEN_5"
mpcli tx marketplace bid TOKEN_5 ${buyerBeneficiary} 50token --from user2 -y <<< '12345678'
sleep 5

echo "Finish TOKEN_5 auction"
mpcli tx marketplace finish_auction TOKEN_5 --from user1 -y <<< '12345678'
sleep 5

# >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

echo "Put TOKEN_6 on auction"
mpcli tx marketplace put_on_auction TOKEN_6 10token ${sellerBeneficiary} 10m --buyout 100token --from user1 -y <<< '12345678'
sleep 5

echo "Bid on TOKEN_6"
mpcli tx marketplace bid TOKEN_6 ${buyerBeneficiary} 50token --from user2 -y <<< '12345678'
sleep 5

# >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

echo "Offer a price for TOKEN_7"
mpcli tx marketplace offer TOKEN_7 100token ${buyerBeneficiary} --from user2 -y <<< '12345678'
sleep 5

echo "Offer a price for TOKEN_6"
mpcli tx marketplace accept_offer TOKEN_7 [offer_id] [beneficiary] --from user1 -y <<< '12345678'
sleep 5
