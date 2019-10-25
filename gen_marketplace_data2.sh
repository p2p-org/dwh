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
for i in 8 9 10
do
	echo "Minting token $i..."
	mpcli tx nft mint ${denom} "TOKEN_$i" ${user1} --from user1 -y <<< '12345678'
	sleep 5
done


for i in 8 9
do
	echo "Remove TOKEN_$i from market"
    mpcli tx marketplace remove_from_market "TOKEN_$i" --from user1 -y <<< '12345678'
	sleep 5
done

for i in 11 12 13
do
	echo "Minting token $i..."
	mpcli tx nft mint ${denom} "TOKEN_$i" ${user1} --from user1 -y <<< '12345678'
	sleep 5
done

for i in 11 12 13
do
	echo "Put TOKEN_$i on auction"
    mpcli tx marketplace put_on_auction "TOKEN_$i" 100000000000000000000000token ${sellerBeneficiary} 9999m --from user1 -y <<< '12345678'
	sleep 5
done

for i in 11 12
do
    echo "Remove TOKEN_$i from auction"
    mpcli tx marketplace remove_from_auction "TOKEN_$i" --from user1 -y <<< '12345678'
	sleep 5
done

# >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

echo "Offer a price again for TOKEN_8"
mpcli tx marketplace offer TOKEN_14 678token ${buyerBeneficiary} --from user2 -y <<< '12345678'
sleep 5

echo "Offer a price again for TOKEN_8"
mpcli tx marketplace offer TOKEN_14 789token ${buyerBeneficiary} --from user2 -y <<< '12345678'
sleep 5

echo "Offer a price again for TOKEN_12"
mpcli tx marketplace offer TOKEN_15 1567000token ${buyerBeneficiary} --from user2 -y <<< '12345678'
sleep 5

echo "Offer a price again for TOKEN_12"
mpcli tx marketplace offer TOKEN_15 1867000token ${buyerBeneficiary} --from user2 -y <<< '12345678'
sleep 5


echo "Put TOKEN_12 on auction"
mpcli tx marketplace put_on_auction TOKEN_14 13token ${sellerBeneficiary} 100m --buyout 196token --from user1 -y <<< '12345678'
sleep 5

echo "Put TOKEN_13 on auction"
mpcli tx marketplace put_on_auction TOKEN_15 41token ${sellerBeneficiary} 100m --buyout 230token --from user1 -y <<< '12345678'
sleep 5

for i in 14 15 16
do
	echo "Minting token $i..."
	mpcli tx nft mint ${denom} "TOKEN_$i" ${user2} --from user2 -y <<< '12345678'
	sleep 2

	echo "Put TOKEN_$i on market"
    mpcli tx marketplace put_on_market "TOKEN_$i" $((i*13))token ${sellerBeneficiary} --from user2 -y <<< '12345678'
	sleep 2
done
