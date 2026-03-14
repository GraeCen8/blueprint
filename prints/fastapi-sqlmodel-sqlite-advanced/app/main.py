from fastapi import FastAPI

from .api import router as items_router
from .core.config import settings
from .db import init_db

app = FastAPI(title=settings.app_name)


@app.on_event("startup")
def on_startup() -> None:
    init_db()


app.include_router(items_router)
