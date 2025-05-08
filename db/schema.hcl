schema "public" {
}

table "user" {
  schema = schema.public
  column "id" {
    type     = uuid
    null     = false
  }
  column "email" {
    type     = varchar(255)
    null     = false
  }
  column "name" {
    type     = varchar(255)
    null     = false
  }
  column "password" {
    type     = varchar(60)
    null     = false
  }
  column "refresh_token" {
    type     = varchar(1024)
    null     = true
  }
  column "created_at" {
    type     = timestamp
    default  = sql("NOW()")
  }
  column "updated_at" {
    type     = timestamp
    default  = sql("NOW()")
  }

  unique "unique_email" {
    columns = [column.email]
  }
  primary_key {
    columns = [column.id]
  }
}

# enum "game" {
#   schema = schema.public
#   values = ["tic_tac_toc"]
# }

# table "room" {
#   schema = schema.public
#   column "id" {
#     type     = uuid
#     null     = false
#   }
#   column "game_type" {
#     type = game
#     null = false
#   }
# }

# table "room_users" {
#   schema = schema.public
#   column "id" {
#     type     = uuid
#     null     = false
#   }
#   column "room_id" {
#     type     = uuid
#     null = false
#   }
#   column "user_id" {
#     type     = uuid
#     null = false
#   }
# }
