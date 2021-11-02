from flask import request
from werkzeug.exceptions import abort
import jwt


def getuser():
    if not("Authorization" in request.headers):
        abort(401, description="Missing header")
    raw_header = request.headers["Authorization"]
    split_header = raw_header.split(" ")
    if len(split_header) != 2:
        abort(401, description="Wrong header")

    return "none"


def login():
    email = request.form.get("email")
    password = request.form.get("password")
    if email is None:
        abort(400, description="Missing email")
    if password is None:
        abort(400, description="Missing password")
    encoded = jwt.encode({"id": 1}, "secret", algorithm="HS256")
    return "Bearer " + encoded


def register():
    email = request.form.get("email")
    password = request.form.get("password")
    if email is None:
        abort(400, description="Missing email")
    if password is None:
        abort(400, description="Missing password")
    return login()
