import tensorflow as tf
import numpy as np
import scipy.io as sc

from tensorflow.python.tools import saved_model_cli as sm

from main import preprocess

data_dir  = "/home/ubuntu/dataset/"
# TODO : model name have to be parsed from user argument
model_dir = "/home/ubuntu/models/model"
out_dir   = "/home/ubuntu/out/"

tag_set = "serve"
signature_def = "default"

try:
    # TODO: Can we believe this 'input_label' from user-defined function?
    input_data, input_label = preprocess(data_dir)
except NameError: # Case when the user didn't implement 'preprocess()'
    print("error: user didn't implement the function 'preprocess()'.")
    exit(1)

est = []

# Segment input_data into some batches to avoid Out-of-Memory Exception
for i in range(0, len(input_data), 50) :
    feed_dict = {
        "inputs" : input_data[i:i+50]
    }

    # TODO: Maybe we can re-implement this function for future modification.
    sm.run_saved_model_with_feed_dict(model_dir,
                                      tag_set,
                                      signature_def,
                                      feed_dict,
                                      out_dir,
                                      True)

    est += np.load(out_dir + 'outputs.npy').tolist()

# Evaluates the result
# TODO: Make several evaluator per output type as functions.
#       (Currently, only works for classification cases)

est = [ np.argmax(e) for e in est ]
ans = [ a for a in input_label ]
#ans = [ np.argmax(a) for a in input_label ]

scr = [ 1 for e, a in zip(est, ans) if (e == a) ].count(1)

print("correct: %d/%d (%3.1f%%)" % (scr, len(est), 100*float(scr)/len(est)))
