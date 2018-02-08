# The original tensorflow docker includes some python modules below
# : Pillow, h5py, ipykernel, jupyter, matplotlib, numpy, pandas, scipy, sklearn

#FROM tensorflow/tensorflow:latest[-gpu][-py3]
FROM tensorflow/tensorflow:latest-gpu

# Add user
RUN useradd -m -s /bin/bash -N -u 1000 ubuntu
USER ubuntu

# Set work directory
RUN mkdir -p /home/ubuntu/out /home/ubuntu/src/ /home/ubuntu/dat
WORKDIR /home/ubuntu

# Set default command as bash
CMD "/bin/bash"
