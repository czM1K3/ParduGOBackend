def is_integer(string):
	try:
		return int(string)
	except ValueError:
		return None

def is_float(string):
	try:
		return float(string)
	except ValueError:
		return None