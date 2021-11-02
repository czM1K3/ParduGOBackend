from flask import Flask, request, jsonify
from werkzeug.exceptions import abort
import auth
import points
from dotenv import load_dotenv

app = Flask(__name__)


@app.errorhandler(401)
def unauthorized(e):
    return jsonify(error=str(e)), 401


@app.errorhandler(400)
def unauthorized(e):
    return jsonify(error=str(e)), 400


@app.route("/")
def index():
    return "Hello World!!!"


@app.route("/api/login", methods=["post"])
def login():
    bearer = auth.login()
    return jsonify(token=bearer)


@app.route("/api/register", methods=["post"])
def register():
    bearer = auth.register()
    return jsonify(token=bearer)

@app.route("/api/point/create", methods=["post"])
def create_point():
    points.create()
    return jsonify(success=True)

@app.route("/api/point/get")
def get_point():
    return jsonify(points.points())


@app.route("/api/get")
def get():
    user = auth.getuser()
    return "get"


if __name__ == "__main__":
    load_dotenv()
    app.run(debug=True)
