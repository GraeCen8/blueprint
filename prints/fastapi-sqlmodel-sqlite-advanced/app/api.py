from fastapi import APIRouter, Depends, HTTPException
from sqlmodel import Session

from .crud import create_item, get_item, list_items
from .db import get_session
from .schemas import ItemCreate, ItemRead

router = APIRouter(prefix="/items", tags=["items"])


@router.post("", response_model=ItemRead)
def create_item_endpoint(
    payload: ItemCreate, session: Session = Depends(get_session)
) -> ItemRead:
    return create_item(session, payload)


@router.get("", response_model=list[ItemRead])
def list_items_endpoint(
    session: Session = Depends(get_session),
) -> list[ItemRead]:
    return list_items(session)


@router.get("/{item_id}", response_model=ItemRead)
def get_item_endpoint(
    item_id: int, session: Session = Depends(get_session)
) -> ItemRead:
    item = get_item(session, item_id)
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    return item
