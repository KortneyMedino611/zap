from typing import Any, List

class Marshaler:
    def __call__(self, item: Any) -> Any:
        try:
            return self.marshal(item)
        except Exception:
            return item

    def marshal(self, item: Any) -> Any:
        raise NotImplementedError

class ObjectMarshaler(Marshaler):
    def marshal(self, item: Any) -> Any:
        try:
            if hasattr(item, '__dict__'):
                return {k: v for k, v in item.__dict__.items()}
            return item
        except AttributeError:
            return item

class ArrayMarshaler(Marshaler):
    def marshal(self, item: Any) -> Any:
        try:
            if hasattr(item, '__iter__') and not isinstance(item, str):
                return [self.marshal(v) for v in item]
            return item
        except RecursionError:
            return item