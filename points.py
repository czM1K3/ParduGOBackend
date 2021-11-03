from flask import request
from werkzeug.exceptions import abort
from auth import get_user_id
from database import create_point, get_points
from validate import is_float, is_integer

def create():
	name = request.form.get("name")
	description = request.form.get("description")
	latitude = request.form.get("latitude")
	longitude = request.form.get("longitude")
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
	if not is_float(latitude):
		abort(400, description="Latitude must be a float")
	if not is_float(longitude):
		abort(400, description="Longitude must be a float")
	user = get_user_id()
	create_point(user, name, description, latitude, longitude, type)

def points():
	radius = request.form.get("radius")
	latitude = request.form.get("latitude")
	longitude = request.form.get("longitude")
	if radius is None:
		abort(400, description="Missing radius")
	if latitude is None:
		abort(400, description="Missing latitude")
	if longitude is None:
		abort(400, description="Missing longitude")
	if not is_float(radius):
		abort(400, description="Radius must be a float")
	if not is_float(latitude):
		abort(400, description="Latitude must be a float")
	if not is_float(longitude):
		abort(400, description="Longitude must be a float")
	points = get_points(float(latitude), float(longitude), float(radius))

	list = []
	for x in points:
		list.append({"name": x["name"], "description": x["description"], "longitude": x["location"]["coordinates"][0], "latitude": x["location"]["coordinates"][1], "type": x["type"]})

	return list