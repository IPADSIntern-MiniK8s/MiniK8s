from flask import Flask, request
import func

app = Flask(__name__)


@app.route('/', methods=['POST'])
def handle_request():
    # `params` is a dict
    params = request.json
    try:
        func.run(**params)
        return 'OK'
    except Exception as e:
        return str(e)

if __name__ == '__main__':
    app.run(host="0.0.0.0", port=8081, debug=True)
    

    

