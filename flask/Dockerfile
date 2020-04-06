FROM python:2.7-slim

COPY . ./app
WORKDIR /app

RUN pip install -r requirements.txt

CMD FLASK_APP=server.py FLASK_ENV=development flask run --host=0.0.0.0 -p 3001