#!/usr/bin/env bash
set -euo pipefail

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1"
    exit 1
  fi
}

prompt_with_default() {
  local __var_name="$1"
  local prompt="$2"
  local default_value="$3"
  local input
  read -r -p "$prompt [$default_value]: " input
  printf -v "$__var_name" "%s" "${input:-$default_value}"
}

print_json_or_raw() {
  local payload="$1"
  if jq -e . >/dev/null 2>&1 <<<"$payload"; then
    jq . <<<"$payload"
  else
    echo "$payload"
  fi
}

build_register_payload() {
  local name="$1"
  local username="$2"
  local email="$3"
  local password="$4"

  if [[ -n "$email" ]]; then
    jq -nc \
      --arg name "$name" \
      --arg username "$username" \
      --arg email "$email" \
      --arg password "$password" \
      '{name: $name, username: $username, email: $email, password: $password}'
  else
    jq -nc \
      --arg name "$name" \
      --arg username "$username" \
      --arg password "$password" \
      '{name: $name, username: $username, password: $password}'
  fi
}

TOTAL_STEPS=35
STEP=0
step() {
  STEP=$((STEP + 1))
  echo "$STEP/$TOTAL_STEPS — $1"
}

require_cmd curl
require_cmd jq

DEFAULT_BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
prompt_with_default BASE_URL "Base API URL" "$DEFAULT_BASE_URL"

USER1_NAME="User One"
USER1_USERNAME="user_one_$(date +%s)"
USER1_UPDATED_USERNAME="updated_${USER1_USERNAME}"
USER1_EMAIL_DEFAULT="user_one_$(date +%s)@mail.test"
USER1_PASSWORD="password123"
CURRENT_USER1_PASSWORD="$USER1_PASSWORD"

USER2_NAME="User Two"
USER2_USERNAME="user_two_$(date +%s)"
USER2_EMAIL_DEFAULT="user_two_$(date +%s)@mail.test"
USER2_PASSWORD="password123"

prompt_with_default USER1_EMAIL "User 1 email (real or auto-random)" "$USER1_EMAIL_DEFAULT"
prompt_with_default USER2_EMAIL "User 2 email (real or auto-random)" "$USER2_EMAIL_DEFAULT"

step "Register user 1"
REGISTER1_PAYLOAD=$(build_register_payload "$USER1_NAME" "$USER1_USERNAME" "$USER1_EMAIL" "$USER1_PASSWORD")
REGISTER1=$(curl -sS -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "$REGISTER1_PAYLOAD")
print_json_or_raw "$REGISTER1"

ACCESS_TOKEN=$(jq -er '.access_token' <<<"$REGISTER1")
REFRESH_TOKEN=$(jq -er '.refresh_token' <<<"$REGISTER1")
USER1_ID=$(jq -er '.user.id' <<<"$REGISTER1")

step "Login user 1"
LOGIN1=$(curl -sS -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER1_USERNAME\",\"password\":\"$CURRENT_USER1_PASSWORD\"}")
print_json_or_raw "$LOGIN1"
ACCESS_TOKEN=$(jq -er '.access_token' <<<"$LOGIN1")
REFRESH_TOKEN=$(jq -er '.refresh_token' <<<"$LOGIN1")

step "Refresh tokens"
REFRESH=$(curl -sS -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")
print_json_or_raw "$REFRESH"

step "Verify email (optional)"
read -r -p "Enter VERIFY_TOKEN from email/Redis (press Enter to skip): " VERIFY_TOKEN
if [[ -n "$VERIFY_TOKEN" ]]; then
  VERIFY_RESP=$(curl -sS -X POST "$BASE_URL/auth/verify-email" \
    -H "Content-Type: application/json" \
    -d "{\"token\":\"$VERIFY_TOKEN\"}")
  print_json_or_raw "$VERIFY_RESP"
else
  echo "SKIP (no token)"
fi

step "Forgot password"
FORGOT=$(curl -sS -X POST "$BASE_URL/auth/forgot-password" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER1_EMAIL\"}")
print_json_or_raw "$FORGOT"

step "Set new password (optional)"
read -r -p "Enter RESET_TOKEN from email/Redis (press Enter to skip): " RESET_TOKEN
if [[ -n "$RESET_TOKEN" ]]; then
  SET_NEW_PASS=$(curl -sS -X POST "$BASE_URL/auth/set-new-password" \
    -H "Content-Type: application/json" \
    -d "{\"token\":\"$RESET_TOKEN\",\"new_password\":\"newpassword123\"}")
  print_json_or_raw "$SET_NEW_PASS"
  CURRENT_USER1_PASSWORD="newpassword123"
else
  echo "SKIP"
fi

step "Get current user"
ME=$(curl -sS "$BASE_URL/users/me" -H "Authorization: Bearer $ACCESS_TOKEN")
print_json_or_raw "$ME"

step "Update current user"
UPDATE_ME=$(curl -sS -X PATCH "$BASE_URL/users/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Updated Name\",\"username\":\"$USER1_UPDATED_USERNAME\"}")
print_json_or_raw "$UPDATE_ME"

step "Update avatar (optional)"
read -r -p "Enter AVATAR_PATH (press Enter to skip): " AVATAR_PATH
if [[ -n "$AVATAR_PATH" && -f "$AVATAR_PATH" ]]; then
  UPDATE_AVATAR=$(curl -sS -X PUT "$BASE_URL/users/me/avatar" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -F "avatar=@$AVATAR_PATH")
  print_json_or_raw "$UPDATE_AVATAR"
else
  echo "SKIP"
fi

step "Delete avatar"
read -r -p "Delete avatar now? [y/N]: " RUN_DELETE_AVATAR
if [[ "$RUN_DELETE_AVATAR" =~ ^[Yy]$ ]]; then
  DELETE_AVATAR=$(curl -sS -X DELETE "$BASE_URL/users/me/avatar" \
    -H "Authorization: Bearer $ACCESS_TOKEN")
  print_json_or_raw "$DELETE_AVATAR"
else
  echo "SKIP"
fi

step "Update current password"
UPDATE_PASSWORD=$(curl -sS -X PATCH "$BASE_URL/users/me/update-password" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"current_password\":\"$CURRENT_USER1_PASSWORD\",\"new_password\":\"$USER1_PASSWORD\"}")
print_json_or_raw "$UPDATE_PASSWORD"
CURRENT_USER1_PASSWORD="$USER1_PASSWORD"

step "Get user by ID"
GET_USER_BY_ID=$(curl -sS "$BASE_URL/users/$USER1_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
print_json_or_raw "$GET_USER_BY_ID"

step "Create list"
LIST=$(curl -sS -X POST "$BASE_URL/lists" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"My wishlist","notes":"notes"}')
print_json_or_raw "$LIST"
LIST_ID=$(jq -er '.id' <<<"$LIST")
SHARE_TOKEN=$(jq -er '.share_token' <<<"$LIST")

step "Get current user lists"
GET_LISTS=$(curl -sS "$BASE_URL/lists" -H "Authorization: Bearer $ACCESS_TOKEN")
print_json_or_raw "$GET_LISTS"

step "Get list by ID"
GET_LIST_BY_ID=$(curl -sS "$BASE_URL/lists/$LIST_ID" -H "Authorization: Bearer $ACCESS_TOKEN")
print_json_or_raw "$GET_LIST_BY_ID"

step "Update list"
curl -sS -X PATCH "$BASE_URL/lists/$LIST_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated list","is_public":true}' -i

step "Rotate shared link"
ROTATE=$(curl -sS -X POST "$BASE_URL/lists/$LIST_ID/rotate-share-link" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
print_json_or_raw "$ROTATE"
SHARE_TOKEN=$(jq -er '.share_token' <<<"$ROTATE")

step "Get public lists by user ID"
GET_PUBLIC_LISTS=$(curl -sS "$BASE_URL/users/$USER1_ID/lists" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
print_json_or_raw "$GET_PUBLIC_LISTS"

step "Get list by shared link (guest)"
SHARED_GUEST=$(curl -sS "$BASE_URL/lists/shared/$SHARE_TOKEN")
print_json_or_raw "$SHARED_GUEST"

step "Get list by shared link (authorized)"
SHARED_AUTH=$(curl -sS "$BASE_URL/lists/shared/$SHARE_TOKEN" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
print_json_or_raw "$SHARED_AUTH"

step "Create wish"
WISH=$(curl -sS -X POST "$BASE_URL/lists/$LIST_ID/wishes" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Steam Deck","notes":"512GB","price":59900,"currency":"RUB"}')
print_json_or_raw "$WISH"
WISH_ID=$(jq -er '.id' <<<"$WISH")

step "Update wish"
curl -sS -X PATCH "$BASE_URL/lists/$LIST_ID/wishes/$WISH_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"notes":"OLED","price":69900}' -i

step "Register user 2"
REGISTER2_PAYLOAD=$(build_register_payload "$USER2_NAME" "$USER2_USERNAME" "$USER2_EMAIL" "$USER2_PASSWORD")
REGISTER2=$(curl -sS -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "$REGISTER2_PAYLOAD")
print_json_or_raw "$REGISTER2"
ACCESS_TOKEN_2=$(jq -er '.access_token' <<<"$REGISTER2")
REFRESH_TOKEN_2=$(jq -er '.refresh_token' <<<"$REGISTER2")
USER2_ID=$(jq -er '.user.id' <<<"$REGISTER2")

step "Reserve wish by user 2"
curl -sS -X POST "$BASE_URL/lists/$LIST_ID/wishes/$WISH_ID/reserve" \
  -H "Authorization: Bearer $ACCESS_TOKEN_2" -i

step "Release wish by user 2"
curl -sS -X DELETE "$BASE_URL/lists/$LIST_ID/wishes/$WISH_ID/reserve" \
  -H "Authorization: Bearer $ACCESS_TOKEN_2" -i

step "Delete wish by owner"
curl -sS -X DELETE "$BASE_URL/lists/$LIST_ID/wishes/$WISH_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" -i

step "Logout user 1"
LOGOUT1=$(curl -sS -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")
print_json_or_raw "$LOGOUT1"

step "Logout user 2"
LOGOUT2=$(curl -sS -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN_2" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN_2\"}")
print_json_or_raw "$LOGOUT2"

step "Re-login users for optional destructive endpoints"
LOGIN1_AGAIN=$(curl -sS -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER1_UPDATED_USERNAME\",\"password\":\"$CURRENT_USER1_PASSWORD\"}")
print_json_or_raw "$LOGIN1_AGAIN"
ACCESS_TOKEN=$(jq -er '.access_token' <<<"$LOGIN1_AGAIN")

LOGIN2_AGAIN=$(curl -sS -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER2_USERNAME\",\"password\":\"$USER2_PASSWORD\"}")
print_json_or_raw "$LOGIN2_AGAIN"
ACCESS_TOKEN_2=$(jq -er '.access_token' <<<"$LOGIN2_AGAIN")

step "Delete list (optional)"
read -r -p "Run DELETE /lists/$LIST_ID ? [y/N]: " RUN_DELETE_LIST
if [[ "$RUN_DELETE_LIST" =~ ^[Yy]$ ]]; then
  curl -sS -X DELETE "$BASE_URL/lists/$LIST_ID" \
    -H "Authorization: Bearer $ACCESS_TOKEN" -i
else
  echo "SKIP"
fi

step "Delete current user 1 (optional)"
read -r -p "Run DELETE /users/me for user1 ? [y/N]: " RUN_DELETE_USER1
if [[ "$RUN_DELETE_USER1" =~ ^[Yy]$ ]]; then
  DELETE_USER1=$(curl -sS -X DELETE "$BASE_URL/users/me" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"password\":\"$CURRENT_USER1_PASSWORD\"}")
  print_json_or_raw "$DELETE_USER1"
else
  echo "SKIP"
fi

step "Delete current user 2 (optional)"
read -r -p "Run DELETE /users/me for user2 ? [y/N]: " RUN_DELETE_USER2
if [[ "$RUN_DELETE_USER2" =~ ^[Yy]$ ]]; then
  DELETE_USER2=$(curl -sS -X DELETE "$BASE_URL/users/me" \
    -H "Authorization: Bearer $ACCESS_TOKEN_2" \
    -H "Content-Type: application/json" \
    -d "{\"password\":\"$USER2_PASSWORD\"}")
  print_json_or_raw "$DELETE_USER2"
else
  echo "SKIP"
fi

step "Done"

echo "Tests completed."
step "Summary"
echo "BASE_URL=$BASE_URL"
echo "USER1_ID=$USER1_ID USER2_ID=$USER2_ID LIST_ID=$LIST_ID WISH_ID=$WISH_ID SHARE_TOKEN=$SHARE_TOKEN"
