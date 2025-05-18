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
    type     = varchar(72)
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
