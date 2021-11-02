from flask import request
from werkzeug.exceptions import abort
from auth import get_user_id
from database import create_point, get_points
from validate import is_float, is_integer

def create():
	name = request.form.get("name")
	description = request.form.get("description")
	latitude = is_float(request.form.get("latitude"))
	longitude = is_float(request.form.get("longitude"))
	type = request.form.get("type")
	if name is None:
		abort(400, description="Missing name")
	if description is None:
		abort(400, description="Missing description")
	if latitude is None:
		abort(400, description="Missing latitude")
	if longitude is None:
		abort(400, description="Missing longitude")
	if type is None:
		abort(400, description="Missing type")
	user = get_user_id()
	create_point(user, name, description, latitude, longitude, type)

def points():
	radius = is_integer(request.form.get("radius"))
	latitude = is_float(request.form.get("latitude"))
	longitude = is_float(request.form.get("longitude"))
	if radius is None:
		abort(400, description="Missing radius")
	if latitude is None:
		abort(400, description="Missing latitude")
	if longitude is None:
		abort(400, description="Missing longitude")
	points = get_points(latitude, longitude, radius)

	list = []
	for x in points:
		print(x)
		list.append({"name": x["name"], "description": x["description"], "longitude": x["location"]["coordinates"][0], "latitude": x["location"]["coordinates"][1], "type": x["type"]})

	return list