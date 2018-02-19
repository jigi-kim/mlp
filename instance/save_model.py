import tensorflow as tf

DEFAULT_MODEL_PATH = "/home/ubuntu/models/"
DEFAULT_MODEL_NAME = "model"

def save_model(session, x, y, model_name=DEFAULT_MODEL_NAME):
    builder = tf.saved_model.builder.SavedModelBuilder(DEFAULT_MODEL_PATH + model_name);

    tensor_info_x = tf.saved_model.utils.build_tensor_info(x)
    tensor_info_y = tf.saved_model.utils.build_tensor_info(y)

    signature_def = ( 
        tf.saved_model.signature_def_utils.build_signature_def(
            inputs={'inputs': tensor_info_x},
            outputs={'outputs': tensor_info_y},
            method_name=tf.saved_model.signature_constants.PREDICT_METHOD_NAME
        )
    )

    legacy_init_op = tf.group(tf.tables_initializer(), name='legacy_init_op')

    builder.add_meta_graph_and_variables(
        session,
        [tf.saved_model.tag_constants.SERVING],
        signature_def_map={'default' : signature_def},
        legacy_init_op = legacy_init_op
    )

    builder.save()
