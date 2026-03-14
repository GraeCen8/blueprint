from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    app_name: str = "FastAPI SQLModel Advanced"
    database_url: str = "sqlite:///./app.db"


settings = Settings()
