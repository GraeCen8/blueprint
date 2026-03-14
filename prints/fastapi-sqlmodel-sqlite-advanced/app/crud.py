from sqlmodel import Session, select

from .models import Item
from .schemas import ItemCreate


def create_item(session: Session, data: ItemCreate) -> Item:
    item = Item(name=data.name, description=data.description)
    session.add(item)
    session.commit()
    session.refresh(item)
    return item


def list_items(session: Session) -> list[Item]:
    return session.exec(select(Item)).all()


def get_item(session: Session, item_id: int) -> Item | None:
    return session.get(Item, item_id)
