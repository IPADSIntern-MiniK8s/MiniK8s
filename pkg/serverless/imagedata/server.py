from flask import Flask, request, Response
import func

app = Flask(__name__)


@app.route('/', methods=['POST'])
def handle_request():
    # `params` is a dict
    params = request.json
    headers = {'Content-Type': 'text/plain'}
    try:
        result = func.run(**params)
        response = Response(str(result), headers=headers, status=200)
        return response
    except TypeError as e:
        response = Response("", headers=headers, status=200)
        return response
    except Exception as e:
        response = Response(str(e), headers=headers, status=500)
        return str(e)

if __name__ == '__main__':
    app.run(host="0.0.0.0", port=8081, debug=True)
    

    

