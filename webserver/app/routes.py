from app import app
from drivers.rpc import RPC
from flask import render_template, Response, redirect, url_for, request
from google.protobuf.json_format import MessageToJson
import os
from werkzeug.urls import url_parse

@app.route('/')
@app.route('/index')
def index():
    return render_template('index.html', title='Welcome')

def gen_rtd(rpc):
    while True:
        rtd = rpc.get_rtd()
        yield MessageToJson(rtd)

@app.route('/realtimedata')
def realtimedata():
    return Response(gen_rtd(RPC('localhost:3567')), content_type='application/json')
